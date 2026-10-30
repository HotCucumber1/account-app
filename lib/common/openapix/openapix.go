package openapix

import (
	"log/slog"
	"net/http"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

func ValidationMiddleware(spec *openapi3.T) (func(http.Handler) http.Handler, error) {
	spec.Servers = nil
	router, err := legacy.NewRouter(spec)
	if err != nil {
		return nil, err
	}
	options := &openapi3filter.Options{
		AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, params, err := router.FindRoute(r)
			if err == nil {
				err = openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{
					Request:    r,
					PathParams: params,
					Route:      route,
					Options:    options,
				})
			}
			if err != nil {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func RequestErrorHandler(w http.ResponseWriter, _ *http.Request, _ error) {
	http.Error(w, "invalid request", http.StatusBadRequest)
}

func ResponseErrorHandler(logger *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		logger.ErrorContext(r.Context(), "HTTP request failed", slog.String("path", r.URL.Path), logging.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
