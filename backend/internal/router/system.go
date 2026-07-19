package router

import (
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/handler"
	"github.com/labstack/echo/v4"
)

func registerSystemRoutes(r *echo.Echo, h *handler.Handlers) {
	r.GET("/health", h.Health.CheckHealth)

	r.Static("/static", "static")

	r.GET("/docks", h.OpenAPI.ServeOpenAPIUI)
}
