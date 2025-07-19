package handlers

import (
	"encoding/json"
	"net/http"

	"k8s-autoscale-webapp/models"
)

type StressHandler struct{}

func NewStressHandler() *StressHandler {
	return &StressHandler{}
}

func (h *StressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CPU intensive operation for testing HPA
	iterations := 100000000
	result := 0
	for i := 0; i < iterations; i++ {
		result += i
	}

	response := models.StressTestResponse{
		Message:    "Stress test completed",
		Result:     result,
		Iterations: iterations,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}