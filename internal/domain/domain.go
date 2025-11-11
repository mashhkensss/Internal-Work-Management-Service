package domain

import (
	"internal-work-management-service/internal/domain/note"
	"internal-work-management-service/internal/domain/order"
	"internal-work-management-service/internal/domain/task"
)

type (
	Note      = note.Note
	Task      = task.Task
	Order     = order.Order
	OrderItem = order.Item
)
