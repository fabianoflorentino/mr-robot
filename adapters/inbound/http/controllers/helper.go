package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fabianoflorentino/mr-robot/internal/app/controller"
)

func loadControllerConfig(w http.ResponseWriter) *controller.ConfigManager {
	cfg := controller.NewConfigManager()
	if err := cfg.LoadConfig(); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Error loading configuration")
		return nil
	}

	return cfg
}

// Helper function to write error responses
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string, details ...string) {
	response := map[string]any{"error": message}

	if len(details) > 0 {
		response["details"] = details[0]
	}

	writeJSONResponse(w, statusCode, response)
}

// Helper function to write JSON responses
func writeJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	configManager := controller.NewConfigManager()

	if err := configManager.LoadConfig(); err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	cfg := configManager.GetConfig()
	w.Header().Set(cfg.ContentType, cfg.ApplicationJSON)
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		}
	}
}
