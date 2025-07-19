package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"k8s-autoscale-webapp/models"

	"github.com/go-redis/redis/v8"
)

type HealthHandler struct {
	DB  *sql.DB
	RDB *redis.Client
	Ctx context.Context
}

func NewHealthHandler(db *sql.DB, rdb *redis.Client, ctx context.Context) *HealthHandler {
	return &HealthHandler{
		DB:  db,
		RDB: rdb,
		Ctx: ctx,
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbStatus := "connected"
	if err := h.DB.Ping(); err != nil {
		dbStatus = "disconnected"
	}

	redisStatus := "connected"
	if _, err := h.RDB.Ping(h.Ctx).Result(); err != nil {
		redisStatus = "disconnected"
	}

	response := models.HealthResponse{
		Status:    "healthy",
		Database:  dbStatus,
		Redis:     redisStatus,
		Timestamp: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}