package handler

import (
	"errors"
	"net/http"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
)

type HTTPError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func mapErrorToHTTP(err error) *HTTPError {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return &HTTPError{http.StatusNotFound, "not found"}
	case errors.Is(err, model.ErrConflict):
		return &HTTPError{http.StatusConflict, "conflict"}
	case errors.Is(err, model.ErrInvalidID):
		return &HTTPError{http.StatusBadRequest, "invalid id"}
	case errors.Is(err, model.ErrInvalidInput):
		return &HTTPError{http.StatusBadRequest, "invalid input"}
	case errors.Is(err, model.ErrNoAvailableCouriers):
		return &HTTPError{http.StatusConflict, "no available couriers"}
	case errors.Is(err, model.ErrDeliveryNotFound):
		return &HTTPError{http.StatusNotFound, "delivery not found"}
	default:
		return &HTTPError{http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)}
	}
}
