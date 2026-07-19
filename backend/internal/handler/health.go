package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/middleware"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/server"
	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	Handler
}

func NewHealthHandler(s *server.Server) *HealthHandler {
	return &HealthHandler{
		Handler: Handler{server: s},
	}
}

func (h *HealthHandler) CheckHealth(c echo.Context) error {
	start := time.Now()
	logger := middleware.GetLogger(c).With().
		Str("operation", "health_check").
		Logger()

	response := map[string]interface{}{
		"status":      "healthy",
		"timestamp":   time.Now().UTC(),
		"environment": h.server.Config.Primary.Env,
		"checks":      make(map[string]interface{}),
	}

	checks := response["checks"].(map[string]interface{})
	isHealthy := true

	// check database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbStart := time.Now()
	if err := h.server.DB.Pool.Ping(ctx); err != nil {
		checks["database"] = map[string]interface{}{
			"status":        "unhealthy",
			"response_time": time.Since(dbStart).String(),
			"error":         err.Error(),
		}
		isHealthy = false
		logger.Error().Err(err).Dur("response_time", time.Since(dbStart)).Msg("database health check failed")
		if h.server.LoggerService != nil && h.server.LoggerService.GetApplication() != nil {
			h.server.LoggerService.GetApplication().RecordCustomEvent(
				"HealthCheck", map[string]interface{}{
					"check_type":    "database",
					"operation":     "health_check",
					"error_type":    "database_unhealthy",
					"response_time": time.Since(dbStart).Milliseconds(),
					"error":         err.Error(),
				},
			)
		}
	} else {
		checks["database"] = map[string]interface{}{
			"status":        "healthy",
			"response_time": time.Since(dbStart).String(),
		}
		logger.Info().Dur("response_time", time.Since(dbStart)).Msg("database health check passed")
	}

	// database connection metrics are automatically captured by New Relic nrpgx5 integration

	// check Redis connectivity
	if h.server.Redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		redisStart := time.Now()
		if err := h.server.Redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = map[string]interface{}{
				"status":        "unhealthy",
				"response_time": time.Since(redisStart).String(),
				"error":         err.Error(),
			}
			logger.Error().Err(err).Dur("response_time", time.Since(redisStart)).Msg("redis health check failed")
			if h.server.LoggerService != nil && h.server.LoggerService.GetApplication() != nil {
				h.server.LoggerService.GetApplication().RecordCustomEvent(
					"HealthCheckError", map[string]interface{}{
						"check_type":    "redis",
						"operation":     "health_check",
						"error_type":    "redis_unhealthy",
						"response_time": time.Since(redisStart).Milliseconds(),
						"error":         err.Error(),
					},
				)
			}
		} else {
			checks["redis"] = map[string]interface{}{
				"status":        "healthy",
				"response_time": time.Since(redisStart).String(),
			}
			logger.Info().Dur("response_time", time.Since(redisStart)).Msg("redis health check passed")
		}
	}

	// Set overall status
	if !isHealthy {
		response["status"] = "unhealthy"
		logger.Warn().
			Dur("total_duration", time.Since(start)).
			Msg("health check failed")
		if h.server.LoggerService != nil && h.server.LoggerService.GetApplication() != nil {
			h.server.LoggerService.GetApplication().RecordCustomEvent(
				"HealthCheckError", map[string]interface{}{
					"check_type":    "overall",
					"operation":     "health_check",
					"error_type":    "overall_unhealthy",
					"response_time": time.Since(start).Milliseconds(),
				},
			)
		}
		return c.JSON(http.StatusServiceUnavailable, response)
	}

	logger.Info().
		Dur("total_duration", time.Since(start)).
		Msg("health check passed")

	err := c.JSON(http.StatusOK, response)
	if err != nil {
		logger.Error().Err(err).Msg("failed to write health check JSON response")
		if h.server.LoggerService != nil && h.server.LoggerService.GetApplication() != nil {
			h.server.LoggerService.GetApplication().RecordCustomEvent(
				"HealthCheckError", map[string]interface{}{
					"check_type":    "response",
					"operation":     "health_check",
					"error_type":    "json_response_error",
					"error_message": err.Error(),
				},
			)
		}
		return fmt.Errorf("failed to write JSON response: %w", err)
	}

	return nil
}
