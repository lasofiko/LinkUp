package matching

import (
	"slices"

	"github.com/lasofiko/LinkUp/internal/domain"
)

func FilterRestrictedCandidates(
	currentUserID int,
	candidates []domain.Profile,
	likedIDs []int,
	matchedIDs []int,
) []domain.Profile {
	excluded := make(map[int]struct{}, 1+len(likedIDs)+len(matchedIDs))

	excluded[currentUserID] = struct{}{}

	for _, id := range likedIDs {
		excluded[id] = struct{}{}
	}

	for _, id := range matchedIDs {
		excluded[id] = struct{}{}
	}

	result := make([]domain.Profile, 0, len(candidates))

	for _, candidate := range candidates {
		if !candidate.Available {
			continue
		}

		if _, found := excluded[candidate.UserID]; found {
			continue
		}

		candidate.Interests = slices.Clone(candidate.Interests)
		result = append(result, candidate)
	}

	return result
}
