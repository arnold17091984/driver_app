package apperror

import (
	"encoding/json"
	"net/http"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(status int, code, message string) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

var (
	ErrUnauthorized    = New(http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	ErrForbidden       = New(http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
	ErrNotFound        = New(http.StatusNotFound, "NOT_FOUND", "resource not found")
	ErrBadRequest      = New(http.StatusBadRequest, "BAD_REQUEST", "invalid request")
	ErrConflict        = New(http.StatusConflict, "CONFLICT", "resource conflict")
	ErrInternal        = New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	ErrInvalidCredentials = New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid employee ID or password")
)

type ErrorResponse struct {
	Error *AppError `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, err *AppError) {
	WriteJSON(w, err.Status, ErrorResponse{Error: err})
}

func WriteErrorMsg(w http.ResponseWriter, status int, code, message string) {
	WriteError(w, New(status, code, message))
}

func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, data)
}

func WriteCreated(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, data)
}
