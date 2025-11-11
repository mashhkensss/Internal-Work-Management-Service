package orders

import (
	"errors"
	"net/http"

	"internal-work-management-service/internal/domain"
	domainOrder "internal-work-management-service/internal/domain/order"
)

func mapError(err error) (int, error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, errors.New("err order not found")
	case errors.Is(err, domainOrder.ErrEmptyCustomer),
		errors.Is(err, domainOrder.ErrNoItems),
		errors.Is(err, domainOrder.ErrEmptyItemName),
		errors.Is(err, domainOrder.ErrNegativeItemCost),
		errors.Is(err, domainOrder.ErrNegativeTotal),
		errors.Is(err, domainOrder.ErrNegativeDiscount),
		errors.Is(err, domainOrder.ErrDiscountOverflow):
		return http.StatusBadRequest, err
	default:
		return http.StatusInternalServerError, err
	}
}
