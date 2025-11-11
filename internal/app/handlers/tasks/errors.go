package tasks

import (
	"errors"
	"net/http"

	"internal-work-management-service/internal/domain"
	domainTask "internal-work-management-service/internal/domain/task"
)

func mapError(err error) (int, error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, errors.New("err task not found")
	case errors.Is(err, domainTask.ErrEmptyTitle),
		errors.Is(err, domainTask.ErrCreatedAtZero),
		errors.Is(err, domainTask.ErrCompletedAtZero):
		return http.StatusBadRequest, err
	default:
		return http.StatusInternalServerError, err
	}
}
