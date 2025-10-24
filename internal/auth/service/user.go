package service

import "github.com/art-vbst/art-backend/internal/auth/repo"

type UserService struct {
	repo repo.Repo
}

func (s *UserService) Create(email string, password string) error {
	return nil
}
