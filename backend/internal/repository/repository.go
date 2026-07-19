package repository

import "github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/server"

// will contain all the repository logic here
// different database instances can be used here and can be injected into the services layer

type Repositories struct{}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{}
}
