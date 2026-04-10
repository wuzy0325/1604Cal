package http

import (
	"encoding/json"
	"net/http"

	"cal1604/internal/api/dto"
)

func writeSuccess[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(dto.Response[T]{
		Success: true,
		Data:    data,
	})
}
