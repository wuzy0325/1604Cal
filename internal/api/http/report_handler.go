package http

import (
	"net/http"
	"strconv"
	"strings"

	apperrors "cal1604/internal/errors"
	"cal1604/internal/report"
)

type reportTemplateSelection struct {
	Filename string `json:"filename"`
}

func (s *apiServer) reportTemplateSelectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pointsText := strings.TrimSpace(r.URL.Query().Get("points"))
	mode := strings.TrimSpace(r.URL.Query().Get("mode"))
	if pointsText == "" || mode == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	points, err := strconv.Atoi(pointsText)
	if err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	filename, err := report.SelectTemplate(points, mode)
	if err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	writeSuccess(w, http.StatusOK, reportTemplateSelection{Filename: filename})
}
