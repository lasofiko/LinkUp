package usecase

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type discoveryStub struct {
	users      []domain.User
	likedIDs   []int
	matchedIDs []int

	usersErr   error
	likesErr   error
	matchesErr error

	likesUserID   int
	matchesUserID int
}

func (s *discoveryStub) GetAll() ([]domain.User, error) {
	return s.users, s.usersErr
}

func (s *discoveryStub) GetLikedUserIDs(userID int) ([]int, error) {
	s.likesUserID = userID
	return s.likedIDs, s.likesErr
}

func (s *discoveryStub) GetMatchedUserIDs(userID int) ([]int, error) {
	s.matchesUserID = userID
	return s.matchedIDs, s.matchesErr
}

func TestRestrictedDiscoveryServiceDiscover(t *testing.T) {
	stub := &discoveryStub{
		users: []domain.User{
			{ID: 1, Available: true},
			{ID: 2, Available: true, Interests: []string{"Go"}},
			{ID: 3, Available: false},
			{ID: 4, Available: true},
			{ID: 5, Available: true},
		},
		likedIDs:   []int{4},
		matchedIDs: []int{5},
	}

	service := NewRestrictedDiscoveryService(stub, stub, stub)

	got, err := service.Discover(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []domain.User{
		{ID: 2, Available: true, Interests: []string{"Go"}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}

	if stub.likesUserID != 1 || stub.matchesUserID != 1 {
		t.Fatal("service passed an incorrect current user ID")
	}

	got[0].Interests[0] = "Changed"

	if stub.users[1].Interests[0] != "Go" {
		t.Fatal("service result changed source interests")
	}
}

func TestRestrictedDiscoveryServiceEmptyResult(t *testing.T) {
	stub := &discoveryStub{
		users: []domain.User{
			{ID: 1, Available: true},
		},
	}

	service := NewRestrictedDiscoveryService(stub, stub, stub)

	got, err := service.Discover(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil || len(got) != 0 {
		t.Fatalf("expected an empty non-nil slice, got %#v", got)
	}
}

func TestRestrictedDiscoveryServiceSourceErrors(t *testing.T) {
	for _, source := range []string{"users", "likes", "matches"} {
		for _, wrapped := range []bool{false, true} {
			name := fmt.Sprintf("%s/wrapped=%t", source, wrapped)

			t.Run(name, func(t *testing.T) {
				rootErr := errors.New("source failure")
				sourceErr := rootErr

				if wrapped {
					sourceErr = fmt.Errorf("repository: %w", rootErr)
				}

				stub := &discoveryStub{}

				switch source {
				case "users":
					stub.usersErr = sourceErr
				case "likes":
					stub.likesErr = sourceErr
				case "matches":
					stub.matchesErr = sourceErr
				}

				service := NewRestrictedDiscoveryService(
					stub,
					stub,
					stub,
				)

				got, err := service.Discover(1)

				if !errors.Is(err, rootErr) {
					t.Fatalf("expected source error, got %v", err)
				}

				if err == sourceErr {
					t.Fatal("expected service to add error context")
				}

				if got != nil {
					t.Fatalf("expected nil result on error, got %#v", got)
				}
			})
		}
	}
}
