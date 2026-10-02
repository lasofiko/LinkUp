package usecase

import (
	"fmt"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/matching"
)

type DiscoveryUserSource interface {
	GetAll() ([]domain.User, error)
}

type DiscoveryLikeSource interface {
	GetLikedUserIDs(userID int) ([]int, error)
}

type DiscoveryMatchSource interface {
	GetMatchedUserIDs(userID int) ([]int, error)
}

type RestrictedDiscoveryService struct {
	users   DiscoveryUserSource
	likes   DiscoveryLikeSource
	matches DiscoveryMatchSource
}

func NewRestrictedDiscoveryService(
	users DiscoveryUserSource,
	likes DiscoveryLikeSource,
	matches DiscoveryMatchSource,
) *RestrictedDiscoveryService {
	return &RestrictedDiscoveryService{
		users:   users,
		likes:   likes,
		matches: matches,
	}
}

func (s *RestrictedDiscoveryService) Discover(
	currentUserID int,
) ([]domain.User, error) {
	candidates, err := s.users.GetAll()
	if err != nil {
		return nil, fmt.Errorf("get discovery candidates: %w", err)
	}

	likedIDs, err := s.likes.GetLikedUserIDs(currentUserID)
	if err != nil {
		return nil, fmt.Errorf("get liked users: %w", err)
	}

	matchedIDs, err := s.matches.GetMatchedUserIDs(currentUserID)
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
