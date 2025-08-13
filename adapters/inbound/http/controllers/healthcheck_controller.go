package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/fabianoflorentino/mr-robot/internal/app/controller"
)

type HealthCheckController struct{}

func NewHealthCheckController() *HealthCheckController {
	return &HealthCheckController{}
}

func (h *HealthCheckController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	configManager := controller.NewConfigManager()
	err := configManager.LoadConfig()
	if err != nil {
		http.Error(w, "Error loading configuration", http.StatusInternalServerError)
		return
	}
	cfg := configManager.GetConfig()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]any{"service": cfg.HostName, "status": http.StatusOK}

	w.Header().Set(cfg.ContentType, cfg.ApplicationJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}
