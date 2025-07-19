package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"k8s-autoscale-webapp/config"
	"k8s-autoscale-webapp/handlers"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg := config.Load()
	ctx := context.Background()

	// Initialize database
	db, err := initDB(cfg.DatabaseConfig)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize Redis
	rdb, err := initRedis(cfg.RedisConfig, ctx)
	if err != nil {
		log.Printf("Redis connection failed: %v", err)
	} else {
		log.Println("Redis connected successfully")
		defer rdb.Close()
	}

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db, rdb, ctx)
	userHandler := handlers.NewUserHandler(db, rdb, ctx)
	stressHandler := handlers.NewStressHandler()

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Health check endpoint
	mux.Handle("GET /health", healthHandler)

	// User endpoints using Go 1.22+ pattern matching
	mux.HandleFunc("GET /api/users", userHandler.GetUsers)
	mux.HandleFunc("POST /api/users", userHandler.CreateUser)
	mux.HandleFunc("GET /api/users/{id}", userHandler.GetUser)

	// Stress test endpoint
	mux.Handle("GET /api/stress", stressHandler)

	// Wrap with CORS middleware
	handler := handlers.CORSMiddleware(mux)

	log.Printf("Server starting on port %s...", cfg.ServerConfig.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerConfig.Port, handler))
}

func initDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Printf("Database connection failed: %v", err)
		return db, nil // Return db anyway for health checks
	}

	// Create users table if not exists
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100),
		email VARCHAR(100) UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")
	return db, nil
}

func initRedis(cfg config.RedisConfig, ctx context.Context) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := rdb.Ping(ctx).Result()
	return rdb, err
}