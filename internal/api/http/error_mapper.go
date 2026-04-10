package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"cal1604/internal/api/dto"
	apperrors "cal1604/internal/errors"
)

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "internal error"

	if errors.Is(err, apperrors.ErrUnitMismatch) {
		status = http.StatusBadRequest
		code = "UNIT_MISMATCH"
		message = "device units are not consistent"
	} else if errors.Is(err, apperrors.ErrInvalidArgument) {
		status = http.StatusBadRequest
		code = "INVALID_ARGUMENT"
		message = "invalid request argument"
	} else if errors.Is(err, apperrors.ErrNotFound) {
		status = http.StatusNotFound
		code = "NOT_FOUND"
		message = "resource not found"
	} else if errors.Is(err, apperrors.ErrInvalidStateTransition) {
		status = http.StatusConflict
		code = "INVALID_STATE_TRANSITION"
		message = "invalid session state transition"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(dto.Response[any]{
		Success: false,
		Code:    code,
		Message: message,
	})
}
