package handler

import (
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/server"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
	}
}
