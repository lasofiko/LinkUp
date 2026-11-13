package usecase

import (
	"context"
	"fmt"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/matching"
)

type RestrictedDiscoveryProfileSource interface {
	ListProfiles(ctx context.Context) ([]domain.Profile, error)
}

type RestrictedDiscoveryLikeSource interface {
	GetLikedUserIDs(ctx context.Context, userID int) ([]int, error)
}

type RestrictedDiscoveryMatchSource interface {
	GetMatchedUserIDs(ctx context.Context, userID int) ([]int, error)
}

type RestrictedDiscoveryService struct {
	profiles RestrictedDiscoveryProfileSource
	likes    RestrictedDiscoveryLikeSource
	matches  RestrictedDiscoveryMatchSource
}

func NewRestrictedDiscoveryService(
	profiles RestrictedDiscoveryProfileSource,
	likes RestrictedDiscoveryLikeSource,
	matches RestrictedDiscoveryMatchSource,
) *RestrictedDiscoveryService {
	return &RestrictedDiscoveryService{
		profiles: profiles,
		likes:    likes,
		matches:  matches,
	}
}

func (s *RestrictedDiscoveryService) Discover(
	ctx context.Context,
	currentUserID int,
) ([]domain.Profile, error) {
	candidates, err := s.profiles.ListProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("get discovery profiles: %w", err)
	}

	likedIDs, err := s.likes.GetLikedUserIDs(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("get liked users: %w", err)
	}

	matchedIDs, err := s.matches.GetMatchedUserIDs(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("get matched users: %w", err)
	}

	return matching.FilterRestrictedCandidates(
		currentUserID,
		candidates,
		likedIDs,
		matchedIDs,
	), nil
}
