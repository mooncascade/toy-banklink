package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type httpError struct {
	Message string `json:"message"`
}

// HTTPError outputs json encoded error message to response writer
func HTTPError(message string, httpCode int, w http.ResponseWriter) {
	httpError{
		Message: message,
	}.writeError(httpCode, w)
}

func setHeaders(httpCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
}

func (err httpError) writeError(httpCode int, w http.ResponseWriter) {
	setHeaders(httpCode, w)

	if encodeErr := json.NewEncoder(w).Encode(err); encodeErr != nil {
		log.Println("Unable to encode error message {0}", encodeErr)
	}

	// If error code is not 400-499, then log the error to server aswell
	if !(httpCode >= 400 && httpCode <= 499) {
		log.Println("Error code: "+string(httpCode), err)
	}
}
