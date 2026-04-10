package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "cal1604/internal/errors"
)

func TestWriteErrorResponse(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, apperrors.ErrUnitMismatch)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "UNIT_MISMATCH") {
		t.Fatalf("expected UNIT_MISMATCH in body, got %q", body)
	}
}
