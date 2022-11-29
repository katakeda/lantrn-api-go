package services

import (
	"fmt"

	"github.com/katakeda/lantrn-api-go/repositories"
)

type Service struct {
	repo repositories.IRepository
}

func NewService(repo repositories.IRepository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is required to start a new service")
	}

	return &Service{
		repo: repo,
	}, nil
}
