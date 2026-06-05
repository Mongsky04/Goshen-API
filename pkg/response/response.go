package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload) //nolint:errcheck
}

func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Response{Success: true, Data: data})
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Response{Success: true, Data: data})
}

func BadRequest(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusBadRequest, Response{Success: false, Error: msg})
}

func Unauthorized(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusUnauthorized, Response{Success: false, Error: msg})
}

func NotFound(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusNotFound, Response{Success: false, Error: msg})
}

func InternalError(w http.ResponseWriter) {
	JSON(w, http.StatusInternalServerError, Response{Success: false, Error: "internal server error"})
}
