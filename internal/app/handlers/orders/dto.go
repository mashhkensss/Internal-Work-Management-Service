package orders

import (
	"time"

	"internal-work-management-service/internal/domain"
	domainOrder "internal-work-management-service/internal/domain/order"
)

type createRequest struct {
	Customer string         `json:"customer"`
	Items    []orderItemDTO `json:"items"`
}

type orderItemDTO struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type response struct {
	ID        int64          `json:"id"`
	Customer  string         `json:"customer"`
	Items     []orderItemDTO `json:"items"`
	Total     float64        `json:"total"`
	Discount  float64        `json:"discount"`
	CreatedAt string         `json:"created_at"`
}

func newOrderResponse(o domain.Order) response {
	items := make([]orderItemDTO, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, orderItemDTO{
			Name:  item.Name,
			Price: item.Price,
		})
	}
	return response{
		ID:        o.ID,
		Customer:  o.Customer,
		Items:     items,
		Total:     o.Total,
		Discount:  o.Discount,
		CreatedAt: o.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func mapItems(items []orderItemDTO) []domainOrder.Item {
	res := make([]domainOrder.Item, 0, len(items))
	for _, in := range items {
		res = append(res, domainOrder.Item{
			Name:  in.Name,
			Price: in.Price,
		})
	}
	return res
}
