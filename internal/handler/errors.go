package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolationCode = "23505"

type httpError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

func (e *httpError) Error() string {
	return e.Message
}

var (
	ErrInvalidInput = &httpError{http.StatusBadRequest, "invalid input"}
	ErrNotFound     = &httpError{http.StatusNotFound, "not found"}
	ErrConflict     = &httpError{http.StatusConflict, "conflict"}
	ErrInvalidID    = &httpError{http.StatusBadRequest, "invalid id"}
	ErrInternal     = &httpError{http.StatusInternalServerError, "internal error"}
)

func (h *courierHandler) writeError(w http.ResponseWriter, err error) {
	var httpErr *httpError
	if errors.As(err, &httpErr) {
		h.writeJSON(w, httpErr.Code, httpErr)
		return
	}
	log.Println("Internal error:", err)
	h.writeJSON(w, http.StatusInternalServerError, ErrInternal)
}

func (h *courierHandler) handleDBError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
		h.writeError(w, ErrConflict)
		return
	}
	h.writeError(w, err)
}
