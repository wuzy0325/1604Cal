package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"cal1604/internal/api/dto"
	apperrors "cal1604/internal/errors"
)

func writeSuccess[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(dto.Response[T]{
		Success: true,
		Data:    data,
	})
}

// decodeJSON 从请求体解码 JSON 到 T，未知字段报错返回 ErrInvalidArgument。
func decodeJSON[T any](r *http.Request) (T, error) {
	var v T
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&v); err != nil {
		return v, errors.Join(apperrors.ErrInvalidArgument, err)
	}
	return v, nil
}
