package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type ProfileStore interface {
	GetProfile(ctx context.Context, userID int) (domain.Profile, error)
	UpdateProfile(ctx context.Context, userID int, expectedVersion int64, city string, interests []domain.Interest) (domain.Profile, error)
}

type UpdateProfileInput struct {
	ExpectedVersion int64
	City            string
	Interests       []string
}

type ProfileService struct {
	store ProfileStore
}

func NewProfileService(store ProfileStore) *ProfileService {
	return &ProfileService{store: store}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID int) (domain.Profile, error) {
	if err := ctx.Err(); err != nil {
		return domain.Profile{}, err
	}
	if userID <= 0 {
		return domain.Profile{}, domain.ErrInvalidProfileEdit
	}
	profile, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("get profile %d: %w", userID, err)
	}
	return profile, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, actorID, userID int, input UpdateProfileInput) (domain.Profile, error) {
	if err := ctx.Err(); err != nil {
		return domain.Profile{}, err
	}
	if actorID <= 0 || userID <= 0 {
		return domain.Profile{}, domain.ErrInvalidProfileEdit
	}
	if actorID != userID {
		return domain.Profile{}, domain.ErrProfileForbidden
	}

	city := strings.TrimSpace(input.City)
	if input.ExpectedVersion < 1 || city == "" || input.Interests == nil {
		return domain.Profile{}, domain.ErrInvalidProfileEdit
	}

	interests := make([]domain.Interest, 0, len(input.Interests))
	seen := make(map[string]struct{}, len(input.Interests))
	for _, raw := range input.Interests {
		interest, err := domain.NewInterest(raw)
		if err != nil {
			return domain.Profile{}, fmt.Errorf("%w: %w", domain.ErrInvalidProfileEdit, err)
		}
		normalized := strings.ToLower(string(interest))
		if _, exists := seen[normalized]; exists {
			return domain.Profile{}, domain.ErrInvalidProfileEdit
		}
		seen[normalized] = struct{}{}
		interests = append(interests, interest)
	}

	updated, err := s.store.UpdateProfile(ctx, userID, input.ExpectedVersion, city, interests)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("update profile %d: %w", userID, err)
	}
	return updated, nil
}
