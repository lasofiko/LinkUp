package memory

import (
	"errors"

	"github.com/lasofiko/LinkUp/internal/domain"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository struct {
	users map[int]domain.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[int]domain.User),
	}
}

func (r *UserRepository) Create(user domain.User) error {
	if _, exists := r.users[user.ID]; exists {
		return ErrUserAlreadyExists
	}

	r.users[user.ID] = user

	return nil
}

func (r *UserRepository) GetByID(id int) (domain.User, error) {
	user, exists := r.users[id]

	if !exists {
		return domain.User{}, ErrUserNotFound
	}

	return user, nil
}

func (r *UserRepository) GetAll() ([]domain.User, error) {
	users := make([]domain.User, 0, len(r.users))

	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}