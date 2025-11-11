package handlers_test

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func withRouteParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}
