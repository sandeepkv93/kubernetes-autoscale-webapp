package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"k8s-autoscale-webapp/models"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	DB  *sql.DB
	RDB *redis.Client
	Ctx context.Context
}

func NewUserHandler(db *sql.DB, rdb *redis.Client, ctx context.Context) *UserHandler {
	return &UserHandler{
		DB:  db,
		RDB: rdb,
		Ctx: ctx,
	}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cacheKey := "users:all"
	cachedUsers, err := h.RDB.Get(h.Ctx, cacheKey).Result()
	if err == nil {
		w.Write([]byte(cachedUsers))
		return
	}

	rows, err := h.DB.Query("SELECT id, name, email, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	usersJSON, _ := json.Marshal(users)
	h.RDB.Set(h.Ctx, cacheKey, usersJSON, 5*time.Minute)

	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User
	err := h.DB.QueryRow(
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, created_at",
		req.Name, req.Email).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.Name = req.Name
	user.Email = req.Email

	// Invalidate cache
	h.RDB.Del(h.Ctx, "users:all")

	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("user:%d", id)
	cachedUser, err := h.RDB.Get(h.Ctx, cacheKey).Result()
	if err == nil {
		w.Write([]byte(cachedUser))
		return
	}

	var user models.User
	err = h.DB.QueryRow("SELECT id, name, email, created_at FROM users WHERE id = $1", id).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	userJSON, _ := json.Marshal(user)
	h.RDB.Set(h.Ctx, cacheKey, userJSON, 5*time.Minute)

	json.NewEncoder(w).Encode(user)
}