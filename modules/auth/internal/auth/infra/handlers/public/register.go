package public

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/distributed-programming-2026/lib/common/openapix"
	runtimehttp "github.com/distributed-programming-2026/lib/runtime/http"

	"auth/api/server/publicapi"
)

func Register(server publicapi.StrictServerInterface, logger *slog.Logger) (runtimehttp.RegisterFunc, error) {
	spec, err := publicapi.GetSwagger()
	if err != nil {
		return nil, err
	}
	validator, err := openapix.ValidationMiddleware(spec)
	if err != nil {
		return nil, err
	}
	return func(router chi.Router) error {
		router.Group(func(router chi.Router) {
			router.Use(openapix.ClientInfo, validator)
			strict := publicapi.NewStrictHandlerWithOptions(server, nil, publicapi.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  openapix.RequestErrorHandler,
				ResponseErrorHandlerFunc: openapix.ResponseErrorHandler(logger),
			})
			publicapi.HandlerFromMux(strict, router)
		})
		return nil
	}, nil
}
