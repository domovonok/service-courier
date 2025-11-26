package handler

import (
	"errors"
	"net/http"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type httpError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

func (e *httpError) Error() string {
	return e.Message
}

func mapErrorToHTTP(err error) *httpError {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return &httpError{http.StatusNotFound, "not found"}
	case errors.Is(err, model.ErrConflict):
		return &httpError{http.StatusConflict, "conflict"}
	case errors.Is(err, model.ErrInvalidID):
		return &httpError{http.StatusBadRequest, "invalid id"}
	case errors.Is(err, model.ErrInvalidInput):
		return &httpError{http.StatusBadRequest, "invalid input"}
	case errors.Is(err, model.ErrNoAvailableCouriers):
		return &httpError{http.StatusConflict, "no available couriers"}
	case errors.Is(err, model.ErrDeliveryNotFound):
		return &httpError{http.StatusNotFound, "delivery not found"}
	default:
		return &httpError{http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)}
	}
}
