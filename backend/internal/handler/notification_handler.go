package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/pkg/apperror"
)

type NotificationHandler struct {
	userRepo userRepository
}

func NewNotificationHandler(userRepo userRepository) *NotificationHandler {
	return &NotificationHandler{userRepo: userRepo}
}

func (h *NotificationHandler) UpdateFCMToken(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req dto.UpdateFCMTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.userRepo.UpdateFCMToken(r.Context(), claims.UserID, req.Token); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
