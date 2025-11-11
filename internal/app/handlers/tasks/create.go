package tasks

import (
	"net/http"

	"internal-work-management-service/internal/app/handlers"
	taskService "internal-work-management-service/internal/service/task"
)

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createRequest
	if err := handlers.DecodeJSON(w, r, &req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.Service.Create(ctx, taskService.CreateTask{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		handlers.HandleError(w, err, mapError)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, newTaskResponse(task))
}
