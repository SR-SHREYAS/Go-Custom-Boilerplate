package middleware

import (
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/server"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Middleware struct {
	Global          *GlobalMiddleware
	Auth            *AuthMiddleware
	ContextEnhancer *ContextEnhancer
	Tracing         *TracingMiddleware
	RateLimit       *RateLimitMiddleware
}

func NewMiddleware(s *server.Server) *Middleware {
	// Get New Relic application instance from server
	var nrApp *newrelic.Application
	if s.LoggerService != nil {
		nrApp = s.LoggerService.GetApplication()
	}

	return &Middleware{
		Global:          NewGlobalMiddleware(s),
		Auth:            NewAuthMiddleware(s),
		ContextEnhancer: NewContextEnhancer(s),
		Tracing:         NewTracingMiddleware(s, nrApp),
		RateLimit:       NewRateLimitMiddleware(s),
	}
}
