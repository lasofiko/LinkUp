package usecase

import (
	"context"
	"fmt"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type EligibilityUserReader interface {
	GetByID(id int) (domain.User, error)
}

type EligibilityProfileReader interface {
	GetProfile(ctx context.Context, userID int) (domain.Profile, error)
}

type EligibilityLikeReader interface {
	ExistsLike(ctx context.Context, from, to int) (bool, error)
}

type CandidateEligibilityService struct {
	users    EligibilityUserReader
	profiles EligibilityProfileReader
	likes    EligibilityLikeReader
}

func NewCandidateEligibilityService(
	users EligibilityUserReader,
	profiles EligibilityProfileReader,
	likes EligibilityLikeReader,
) *CandidateEligibilityService {
	return &CandidateEligibilityService{
		users:    users,
		profiles: profiles,
		likes:    likes,
	}
}

func (s *CandidateEligibilityService) Check(
	ctx context.Context,
	userID int,
	candidateID int,
) (domain.EligibilityStatus, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return "", fmt.Errorf("get user %d: %w", userID, err)
	}

	candidate, err := s.users.GetByID(candidateID)
	if err != nil {
		return "", fmt.Errorf("get candidate %d: %w", candidateID, err)
	}

	profile, err := s.profiles.GetProfile(ctx, candidate.ID)
	if err != nil {
		return "", fmt.Errorf("get candidate profile %d: %w", candidate.ID, err)
	}

	alreadyLiked, err := s.likes.ExistsLike(ctx, user.ID, candidate.ID)
	if err != nil {
		return "", fmt.Errorf(
			"check like %d -> %d: %w",
			user.ID,
			candidate.ID,
			err,
		)
	}

	return domain.CheckCandidateEligibility(
		user.ID,
		profile,
		alreadyLiked,
	), nil
}
