package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/config"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/database"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/handler"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/logger"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/repository"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/router"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/server"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/service"
)

const DefaultContextTimeout = 30

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// Initialize new relic logger service
	loggerService := logger.NewLoggerService(cfg.Observability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	if cfg.Primary.Env != "local" {
		if err := database.Migrate(context.Background(), &log, cfg); err != nil {
			log.Fatal().Err(err).Msg("failed to run migrations")
		}
	}

	// Initialize server
	srv, err := server.New(cfg, &log, loggerService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize server")
	}

	// initialize repositories, services, and handlers
	repos := repository.NewRepositories(srv)
	services, serviceErr := service.NewServices(srv, repos)
	if serviceErr != nil {
		log.Fatal().Err(serviceErr).Msg("failed to initialize services")
	}
	handlers := handler.NewHandlers(srv, services)

	// Initialize router
	r := router.NewRouter(srv, handlers, services)

	// setup Http server
	srv.SetupHTTPServer(r)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	// start server
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// wait for interrupt signal to gracefull shutdown the server
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultContextTimeout*time.Second)

	if err = srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}
	stop()
	cancel()

	log.Info().Msg("server exited properly")

}
