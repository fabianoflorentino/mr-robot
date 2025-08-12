package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

var (
	CONTENT_TYPE     string        = "Content-Type"
	APPLICATION_JSON string        = "application/json"
	HOST_NAME        string        = os.Getenv("HOSTNAME")
	STATUS_OK        int           = http.StatusOK
	TIME             string        = time.Now().Format(time.RFC3339)
	TIME_AFTER       time.Duration = 250 * time.Millisecond
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
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set(CONTENT_TYPE, APPLICATION_JSON)
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		}
	}
}
