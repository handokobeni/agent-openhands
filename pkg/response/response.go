package response

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Success sends a successful response
func Success(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	JSON(w, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// ValidationError sends a validation error response
func ValidationError(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// Unauthorized sends an unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "unauthorized"
	}
	Error(w, http.StatusUnauthorized, message)
}

// InternalError sends an internal server error response
func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal server error")
}
