package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

var (
	CONTENT_TYPE     string = "Content-Type"
	APPLICATION_JSON string = "application/json"
	HOST_NAME        string = os.Getenv("HOSTNAME")
	STATUS_OK        int    = http.StatusOK
	TIME             string = time.Now().Format(time.RFC3339)
)

type HealthCheckController struct{}

func NewHealthCheckController() *HealthCheckController {
	return &HealthCheckController{}
}

func (h *HealthCheckController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]any{"service": HOST_NAME, "time": TIME}

	w.Header().Set(CONTENT_TYPE, APPLICATION_JSON)
	w.WriteHeader(STATUS_OK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}
