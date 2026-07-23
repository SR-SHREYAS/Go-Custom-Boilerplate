package repository

import "github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/server"

type Repositories struct{}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{}
}
