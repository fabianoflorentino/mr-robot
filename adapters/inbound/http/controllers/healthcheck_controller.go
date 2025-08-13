package controllers

import (
	"encoding/json"
	"net/http"
)

type HealthCheckController struct{}

func NewHealthCheckController() *HealthCheckController {
	return &HealthCheckController{}
}

func (h *HealthCheckController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	cfg := loadControllerConfig()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]any{"service": cfg.ControllerConfig.HostName, "time": cfg.ControllerConfig.TimeInfo}

	w.Header().Set(cfg.ControllerConfig.ContentType, cfg.ControllerConfig.ApplicationJSON)
	w.WriteHeader(cfg.ControllerConfig.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}
