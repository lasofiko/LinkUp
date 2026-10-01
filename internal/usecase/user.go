package usecase

import (
	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/repository"
)

type UserUseCase struct {
	repository repository.UserRepository
}

func NewUserUseCase(repository repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		repository: repository,
	}
}

func (u *UserUseCase) CreateUser(user domain.User) error {
	return u.repository.Create(user)
}

func (u *UserUseCase) GetUser(id int) (domain.User, error) {
	return u.repository.GetByID(id)
}

func (u *UserUseCase) GetUsers() ([]domain.User, error) {
	return u.repository.GetAll()
}