package handlers

import (
	"errors"
	"net/http"
)

var internalServiceErr = errors.New("internal service error")

func WriteInternalError(w http.ResponseWriter, err error) {
	if err != nil {
		handlerLogger.Error("msg", "internal_handler_error", "error", err)
	} else {
		handlerLogger.Error("msg", "internal_handler_error")
	}
	WriteError(w, http.StatusInternalServerError, internalServiceErr)
}

type errorMapper func(error) (int, error)

func HandleError(w http.ResponseWriter, err error, mapper errorMapper) {
	status, clientErr := mapper(err)
	if status == http.StatusInternalServerError {
		WriteInternalError(w, err)
		return
	}
	WriteError(w, status, clientErr)
}
