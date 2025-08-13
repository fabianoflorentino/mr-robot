package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/fabianoflorentino/mr-robot/internal/app/controller"
)

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
	err := configManager.LoadConfig()
	if err != nil {
		// Fallback to defaults if config loading fails
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
	} else {
		cfg := configManager.GetConfig()
		w.Header().Set(cfg.ContentType, cfg.ApplicationJSON)
		w.WriteHeader(statusCode)
	}

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		}
	}
}
