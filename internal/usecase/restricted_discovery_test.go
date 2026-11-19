package usecase

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
)

type restrictedDiscoveryStub struct {
	profiles   []domain.Profile
	likedIDs   []int
	matchedIDs []int

	profilesErr error
	likesErr    error
	matchesErr  error

	likesUserID   int
	matchesUserID int

	profilesContext context.Context
	likesContext    context.Context
	matchesContext  context.Context
}

func (s *restrictedDiscoveryStub) ListProfiles(
	ctx context.Context,
) ([]domain.Profile, error) {
	s.profilesContext = ctx
	return s.profiles, s.profilesErr
}

func (s *restrictedDiscoveryStub) GetLikedUserIDs(
	ctx context.Context,
	userID int,
) ([]int, error) {
	s.likesContext = ctx
	s.likesUserID = userID
	return s.likedIDs, s.likesErr
}

func (s *restrictedDiscoveryStub) GetMatchedUserIDs(
	ctx context.Context,
	userID int,
) ([]int, error) {
	s.matchesContext = ctx
	s.matchesUserID = userID
	return s.matchedIDs, s.matchesErr
}

func TestRestrictedDiscoveryServiceDiscover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stub := &restrictedDiscoveryStub{
		profiles: []domain.Profile{
			{UserID: 1, Available: true},
			{
				UserID:    2,
				Available: true,
				Interests: []domain.Interest{"Go"},
			},
			{UserID: 3, Available: false},
			{UserID: 4, Available: true},
			{UserID: 5, Available: true},
		},
		likedIDs:   []int{4},
		matchedIDs: []int{5},
	}

	service := NewRestrictedDiscoveryService(stub, stub, stub)

	got, err := service.Discover(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []domain.Profile{
		{
			UserID:    2,
			Available: true,
			Interests: []domain.Interest{"Go"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}

	if stub.likesUserID != 1 || stub.matchesUserID != 1 {
		t.Fatal("incorrect current user ID passed to sources")
	}

	if stub.profilesContext != ctx ||
		stub.likesContext != ctx ||
		stub.matchesContext != ctx {
		t.Fatal("service did not forward the caller's context")
	}

	got[0].Interests[0] = "Changed"

	if stub.profiles[1].Interests[0] != "Go" {
		t.Fatal("service result changed source interests")
	}
}

func TestRestrictedDiscoveryServiceEmptyResult(t *testing.T) {
	stub := &restrictedDiscoveryStub{
		profiles: []domain.Profile{
			{UserID: 1, Available: true},
		},
	}

	service := NewRestrictedDiscoveryService(stub, stub, stub)

	got, err := service.Discover(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil || len(got) != 0 {
		t.Fatalf("expected an empty non-nil slice, got %#v", got)
	}
}

func TestRestrictedDiscoveryServiceSourceErrors(t *testing.T) {
	for _, source := range []string{"profiles", "likes", "matches"} {
		for _, wrapped := range []bool{false, true} {
			name := fmt.Sprintf("%s/wrapped=%t", source, wrapped)

			t.Run(name, func(t *testing.T) {
				rootErr := errors.New("source failure")
				sourceErr := rootErr

				if wrapped {
					sourceErr = fmt.Errorf("repository: %w", rootErr)
				}

				stub := &restrictedDiscoveryStub{}

				switch source {
				case "profiles":
					stub.profilesErr = sourceErr
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

				got, err := service.Discover(context.Background(), 1)

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

func TestRestrictedDiscoveryServiceUsesSharedStores(t *testing.T) {
	ctx := context.Background()

	profiles := memory.NewProfiles([]domain.Profile{
		{UserID: 1, Available: true},
		{UserID: 2, Available: true},
		{UserID: 3, Available: true},
	})

	likes := memory.NewLikes()
	matches := memory.NewMatches()

	service := NewRestrictedDiscoveryService(profiles, likes, matches)

	before, err := service.Discover(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantBefore := []domain.Profile{
		{UserID: 2, Available: true},
		{UserID: 3, Available: true},
	}

	if !reflect.DeepEqual(before, wantBefore) {
		t.Fatalf("got %#v, want %#v", before, wantBefore)
	}

	likes.Add(1, 2)

	if err := matches.Save(ctx, 3, 1); err != nil {
		t.Fatalf("save match: %v", err)
	}

	after, err := service.Discover(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if after == nil || len(after) != 0 {
		t.Fatalf("expected empty result after store updates, got %#v", after)
	}
}
