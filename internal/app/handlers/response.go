package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxRequestBodyBytes = 1 << 20 // 1 мб

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, "err failed to encode response", http.StatusInternalServerError)
	}
}

func WriteError(w http.ResponseWriter, status int, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	WriteJSON(w, status, ErrorResponse{Error: msg})
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			return errors.New("err invalid json body")
		case errors.As(err, &unmarshalErr):
			return errors.New("err invalid field type")
		default:
			return err
		}
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("err invalid json body")
	}

	return nil
}
