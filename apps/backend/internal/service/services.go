package service

import (
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/lib/job"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/repository"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	return &Services{
		Job:  s.Job,
		Auth: authService,
	}, nil
}
