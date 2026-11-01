package worker

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

const maxBodyBytes = 1 << 16

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write response", "err", err)
	}
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return invalid("body is not valid JSON for this endpoint: %v", err)
	}
	return nil
}

func classify(err error) (int, string) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, ErrExists):
		return http.StatusConflict, "exists"
	case errors.Is(err, ErrNotStopped):
		return http.StatusConflict, "not_stopped"
	case errors.Is(err, ErrNotReady):
		return http.StatusConflict, "not_ready"
	case errors.Is(err, ErrBusy):
		return http.StatusConflict, "busy"
	case errors.Is(err, ErrInvalid):
		return http.StatusBadRequest, "invalid"
	default:
		return http.StatusInternalServerError, "internal"
	}
}

func writeError(w http.ResponseWriter, err error) {
	status, code := classify(err)
	if status == http.StatusInternalServerError {
		slog.Error("request failed", "err", err)
	}
	writeJSON(w, status, errorBody{Error: err.Error(), Code: code})
}
