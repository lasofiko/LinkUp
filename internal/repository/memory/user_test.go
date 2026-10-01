package memory

import (
	"errors"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
)

func TestUserRepositoryCreate(t *testing.T) {
	repo := NewUserRepository()

	user := domain.User{
		ID:    1,
		Name:  "Sofa",
		Email: "sofa@example.com",
	}

	err := repo.Create(user)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	savedUser, err := repo.GetByID(1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if savedUser.ID != user.ID {
		t.Errorf(
			"expected user ID %d, got %d",
			user.ID,
			savedUser.ID,
		)
	}
}
func TestUserRepositoryGetByIDNotFound(t *testing.T) {
	repo := NewUserRepository()

	_, err := repo.GetByID(999)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf(
			"expected ErrUserNotFound, got %v",
			err,
		)
	}
}
func TestUserRepositoryDuplicateUser(t *testing.T) {
	repo := NewUserRepository()

	user := domain.User{
		ID:   1,
		Name: "Sofa",
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = repo.Create(user)

	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf(
			"expected ErrUserAlreadyExists, got %v",
			err,
		)
	}
}
func TestUserRepositoryGetAll(t *testing.T) {
	repo := NewUserRepository()

	user1 := domain.User{
		ID:   1,
		Name: "Sofa",
	}

	user2 := domain.User{
		ID:   2,
		Name: "Arisha",
	}

	if err := repo.Create(user1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.Create(user2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	users, err := repo.GetAll()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(users) != 2 {
		t.Errorf(
			"expected 2 users, got %d",
			len(users),
		)
	}
}