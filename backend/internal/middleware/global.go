package middleware

import (
	"errors"
	"net/http"

	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/errs"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/server"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/sqlerr"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type GlobalMiddleware struct {
	server *server.Server
}

func NewGlobalMiddleware(s *server.Server) *GlobalMiddleware {
	return &GlobalMiddleware{server: s}
}

func (global *GlobalMiddleware) CORS() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: global.server.Config.Server.CORSAllowOrigin,
	})
}

func (global *GlobalMiddleware) RequestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:     true,
		LogStatus:  true,
		LogError:   true,
		LogLatency: true,
		LogHost:    true,
		LogMethod:  true,
		LogURIPath: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			statusCode := v.Status

			//status code is picked up by global err handler
			if v.Error != nil {
				var httpErr *errs.HTTPError
				var echoErr *echo.HTTPError
				if errors.As(v.Error, &httpErr) {
					statusCode = httpErr.Status
				} else if errors.As(v.Error, &echoErr) {
					statusCode = echoErr.Code
				}
			}

			// Get enhanced logger from context
			logger := GetLogger(c)

			var e *zerolog.Event

			switch {
			case statusCode >= 500:
				e = logger.Error().Err(v.Error)
			case statusCode >= 400:
				e = logger.Warn()
			default:
				e = logger.Info()
			}

			// Add request ID if available
			if requestID := GetRequestID(c); requestID != "" {
				e.Str("request_id", requestID)
			}

			// Add user context if available
			if userID := GetUserID(c); userID != "" {
				e = e.Str("user_id", userID)
			}

			e.
				Dur("latency", v.Latency).
				Int("status", statusCode).
				Str("method", v.Method).
				Str("uri", v.URI).
				Str("host", v.Host).
				Str("ip", c.RealIP()).
				Str("user_agent", c.Request().UserAgent()).
				Msg("API")

			return nil
		},
	},
	)
}

func (global *GlobalMiddleware) Recover() echo.MiddlewareFunc {
	return middleware.Recover()
}

func (global *GlobalMiddleware) Secure() echo.MiddlewareFunc {
	return middleware.Secure()
}

// to return a formatted error to client no matter where and what error occurs
func (global *GlobalMiddleware) GlobalErrorHandler(err error, c echo.Context) {
	// first try to handle db errors and convert to appropriate http errors
	originalErr := err

	// try to handle known database errors
	// only for error that haven't already converted to HTTPError

	var httpErr *errs.HTTPError
	if !errors.As(err, &httpErr) {
		var echoErr *echo.HTTPError
		if errors.As(err, &echoErr) {
			if echoErr.Code == http.StatusNotFound {
				err = errs.NewNotFoundError("Route not found", false, nil)
			}
		} else {
			// sqlerr handler will convert db errors
			err = sqlerr.HandleError(err)
		}
	}

	// now process the possibly converted error
	var echoErr *echo.HTTPError
	var status int
	var code string
	var message string
	var fieldErrors []errs.FieldError
	var action *errs.Action

	switch {
	case errors.As(err, &httpErr):
		status = httpErr.Status
		code = httpErr.Code
		message = httpErr.Message
		fieldErrors = httpErr.Errors
		action = httpErr.Action

	case errors.As(err, &echoErr):
		status = echoErr.Code
		code = errs.MakeUpperCaseWithUnderscores(http.StatusText(status))
		if msg, ok := echoErr.Message.(string); ok {
			message = msg
		} else {
			message = http.StatusText(echoErr.Code)
		}
	default:
		status = http.StatusInternalServerError
		code = errs.MakeUpperCaseWithUnderscores(http.StatusText(http.StatusInternalServerError))
		message = http.StatusText(http.StatusInternalServerError)
	}

	// log the original error to help with debugging
	logger := *GetLogger(c)

	logger.Error().Stack().
		Err(originalErr).
		Int("status", status).
		Str("error_code", code).
		Msg(message)

	if !c.Response().Committed {
		_ = c.JSON(status, errs.HTTPError{
			Code:     code,
			Message:  message,
			Status:   status,
			Override: httpErr != nil && httpErr.Override,
			Errors:   fieldErrors,
			Action:   action,
		},
		)
	}
}
