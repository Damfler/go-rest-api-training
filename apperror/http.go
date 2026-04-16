package apperror

import (
	"errors"
	"net/http"
)

func HTTPStatus(err error) int {
	var notFound *NotFoundError
	var validation *ValidationError
	var conflict *ConflictError

	switch {
	case errors.As(err, &notFound):
		return http.StatusNotFound
	case errors.As(err, &validation):
		return http.StatusBadRequest
	case errors.As(err, &conflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
