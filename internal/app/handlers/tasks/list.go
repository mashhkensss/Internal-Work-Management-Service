package tasks

import (
	"errors"
	"net/http"
	"strconv"

	"internal-work-management-service/internal/app/handlers"
	taskService "internal-work-management-service/internal/service/task"
)

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var filter taskService.Filter
	if doneVal := r.URL.Query().Get("done"); doneVal != "" {
		val, err := strconv.ParseBool(doneVal)
		if err != nil {
			handlers.WriteError(w, http.StatusBadRequest, errors.New("err invalid done parameter"))
			return
		}
		filter.Done = &val
	}

	tasksList, err := h.Service.List(ctx, filter)
	if err != nil {
		handlers.WriteInternalError(w, err)
		return
	}

	res := make([]response, 0, len(tasksList))
	for _, t := range tasksList {
		res = append(res, newTaskResponse(t))
	}

	handlers.WriteJSON(w, http.StatusOK, res)
}
