package notes

import (
	"errors"
	"net/http"

	"internal-work-management-service/internal/domain"
	domainNote "internal-work-management-service/internal/domain/note"
)

func mapError(err error) (int, error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, errors.New("err note not found")
	case errors.Is(err, domainNote.ErrEmptyTitle),
		errors.Is(err, domainNote.ErrEmptyContent),
		errors.Is(err, domainNote.ErrCreatedAtZero):
		return http.StatusBadRequest, err
	default:
		return http.StatusInternalServerError, err
	}
}
