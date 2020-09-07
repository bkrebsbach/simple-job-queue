package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

const (
	ErrInvalidInput        = "invalid input"
	ErrInternalServerError = "internal error"
	ErrNotFound            = "not found"
	ErrQueueEmpty          = "queue empty"
)

// ErrorResponse is a simple JSON error response.
type ErrorResponse struct {
	Message string `json:"message"`
}

// WriteJSONStatus is a wrapper for WriteJSONResponse that returns a marshalled JSONStatus blob
func WriteErrorResponse(rw http.ResponseWriter, message string, statusCode int) {
	jsonData, _ := json.Marshal(&ErrorResponse{
		Message: message,
	})

	WriteJSONResponse(rw, statusCode, jsonData)
}

// WriteJSONResponse writes data and status code to the ResponseWriter
func WriteJSONResponse(rw http.ResponseWriter, statusCode int, content []byte) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	if content != nil {
		_, err := rw.Write(content)
		if err != nil {
			log.Error().Err(err).Msg("error writing response")
		}
	}
}
