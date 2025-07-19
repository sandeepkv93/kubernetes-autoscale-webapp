package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Redis     string    `json:"redis"`
	Timestamp time.Time `json:"timestamp"`
}

type StressTestResponse struct {
	Message    string `json:"message"`
	Result     int    `json:"result"`
	Iterations int    `json:"iterations"`
}