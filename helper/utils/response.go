package utils

import (
	"encoding/json"
	"net/http"
)

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	response := map[string]string{"error": message}
	WriteJSON(w, statusCode, response)
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
