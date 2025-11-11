package testutils

import (
	note "internal-work-management-service/internal/domain/note"
	order "internal-work-management-service/internal/domain/order"
	task "internal-work-management-service/internal/domain/task"
)

func NoteWithID(n note.Note, id int64) (note.Note, error) {
	if id <= 0 {
		return note.Note{}, note.ErrInvalidID
	}
	n.ID = id
	return n, nil
}

func OrderWithID(o order.Order, id int64) (order.Order, error) {
	if id <= 0 {
		return order.Order{}, order.ErrInvalidID
	}
	o.ID = id
	return o, nil
}

func TaskWithID(t task.Task, id int64) (task.Task, error) {
	if id <= 0 {
		return task.Task{}, task.ErrInvalidID
	}
	t.ID = id
	return t, nil
}
