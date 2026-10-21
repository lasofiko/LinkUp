package matching

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type StubProfileRepository struct {
	TargetProfile   domain.Profile
	Profiles        []domain.Profile
	GetProfileErr   error
	ListProfilesErr error

	GetCalls  int
	ListCalls int
	GotUserID int
}

func (s *StubProfileRepository) GetProfile(ctx context.Context, userID int) (domain.Profile, error) {
	s.GetCalls++
	s.GotUserID = userID
	if s.GetProfileErr != nil {
		return domain.Profile{}, s.GetProfileErr
	}
	return s.TargetProfile, nil
}

func (s *StubProfileRepository) ListProfiles(ctx context.Context) ([]domain.Profile, error) {
	s.ListCalls++
	if s.ListProfilesErr != nil {
		return nil, s.ListProfilesErr
	}
	return s.Profiles, nil
}

func TestRecommend_Success(t *testing.T) {
	repo := &StubProfileRepository{
		TargetProfile: profile(1, "Москва", true, "go"),
		Profiles: []domain.Profile{
			profile(1, "Москва", true, "go"),
			profile(2, "Москва", true, "go"),
			profile(3, "Казань", true, "go"),
		},
	}
	service := NewDiscoveryService(repo)

	got, err := service.Recommend(context.Background(), 11)

	if err != nil {
		t.Fatalf("unexcpected mistake: %v", err)
	}
	if repo.GotUserID != 11 {
		t.Errorf("ID %d is given, 11 was expected", repo.GotUserID)
	}
	if !slices.Equal(ids(got), []int{2}) {
		t.Errorf("got %v, wanted [2]", ids(got))
	}
}

func TestRecommend_DependencyErrors(t *testing.T) {
	storageDown := errors.New("storage unavailable")

	tests := []struct {
		name          string
		repo          *StubProfileRepository
		wantCause     error
		wantListCalls int
	}{
		{
			name:          "failed to fetch profile because the storage isn't availible",
			repo:          &StubProfileRepository{GetProfileErr: storageDown},
			wantCause:     storageDown,
			wantListCalls: 0,
		},
		{
			name:          "profile wasn't found",
			repo:          &StubProfileRepository{GetProfileErr: domain.ErrProfileNotFound},
			wantCause:     domain.ErrProfileNotFound,
			wantListCalls: 0,
		},
		{
			name: "wan't able to get list of profiles because the storage isn't availible",
			repo: &StubProfileRepository{
				TargetProfile:   profile(1, "Москва", true, "go"),
				ListProfilesErr: storageDown,
			},
			wantCause:     storageDown,
			wantListCalls: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewDiscoveryService(tc.repo).Recommend(context.Background(), 1)

			if err == nil {
				t.Fatal("error was expected, got nil")
			}
			if got != nil {
				t.Fatalf("expected nil, got %v", got)
			}
			if !errors.Is(err, tc.wantCause) {
				t.Fatalf("cause of the error was lost: %v", err)
			}
			if tc.repo.ListCalls != tc.wantListCalls {
				t.Fatalf("ListProfiles was called %d times, %d expected", tc.repo.ListCalls, tc.wantListCalls)
			}
		})
	}
}

func TestRecommend_EmptyResultIsNotError(t *testing.T) {
	repo := &StubProfileRepository{
		TargetProfile: profile(1, "Москва", true, "go"),
		Profiles:      nil,
	}

	got, err := NewDiscoveryService(repo).Recommend(context.Background(), 1)

	if err != nil {
		t.Fatalf("empty res is not a mistake: %v", err)
	}
	if got == nil {
		t.Fatal("empty list was expected, not nil")
	}
	if len(got) != 0 {
		t.Fatalf("empty list was expected, got %v", ids(got))
	}
}
