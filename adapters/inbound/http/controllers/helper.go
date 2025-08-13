package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/fabianoflorentino/mr-robot/config"
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
	cfg := loadControllerConfig()

	w.Header().Set(cfg.ControllerConfig.ContentType, cfg.ControllerConfig.ApplicationJSON)
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		}
	}
}

func loadControllerConfig() *config.AppConfig {
	cfg, err := config.LoadAppConfig()
	if err != nil {
		return nil
	}

	return cfg
}
