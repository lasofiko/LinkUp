package matching

import (
	"context"
	"fmt"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, userID int) (domain.Profile, error)
	ListProfiles(ctx context.Context) ([]domain.Profile, error)
}

type DiscoveryService struct {
	repo ProfileRepository
}

func NewDiscoveryService(repo ProfileRepository) *DiscoveryService {
	return &DiscoveryService{repo: repo}
}

func (s *DiscoveryService) Recommend(ctx context.Context, userID int) ([]domain.Profile, error) {
	target, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target profile: %w", err)
	}

	candidates, err := s.repo.ListProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profiles: %w", err)
	}

	return MatchCandidates(target, candidates), nil
}
