package usecase

import "github.com/lasofiko/LinkUp/internal/domain"

type UserRepository interface {
	Create(user domain.User) error
	GetByID(id int) (domain.User, error)
	GetAll() ([]domain.User, error)
}

type UserUseCase struct {
	repository UserRepository
}

func NewUserUseCase(repository UserRepository) *UserUseCase {
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
