package usecase

import (
	"context"
	"fmt"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/matching"
)

// PairCandidateSource preserves the existing search and ranking rules.
type PairCandidateSource interface {
	Recommend(ctx context.Context, userID int) ([]domain.Profile, error)
}

type PairDiscoveryService struct {
	candidates PairCandidateSource
	pairs      interaction.Reader
}

// NewPairDiscoveryService must receive the same store instance as React.
func NewPairDiscoveryService(candidates PairCandidateSource, pairs interaction.Reader) *PairDiscoveryService {
	return &PairDiscoveryService{candidates: candidates, pairs: pairs}
}

func (s *PairDiscoveryService) Discover(ctx context.Context, userID int) ([]domain.Profile, error) {
	candidates, err := s.candidates.Recommend(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get pair discovery candidates: %w", err)
	}
	return FilterPairCandidates(userID, candidates, s.pairs)
}

// FilterPairCandidates applies Sprint 2 restrictions to the ordered results of
// matching.MatchCandidates. pairs must be the same store used by React.
func FilterPairCandidates(userID int, candidates []domain.Profile, pairs interaction.Reader) ([]domain.Profile, error) {
	result := make([]domain.Profile, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.UserID == userID || !candidate.Available {
			continue
		}
		pair, err := interaction.NewPair(userID, candidate.UserID)
		if err != nil {
			return nil, fmt.Errorf("create discovery pair: %w", err)
		}
		state, found, err := pairs.Load(pair)
		if err != nil {
			return nil, fmt.Errorf("load discovery pair %v: %w", pair, err)
		}
		if found {
			if state.Pair != pair {
				return nil, fmt.Errorf("load discovery pair %v: %w", pair, interaction.ErrStatePair)
			}
			if state.HasLike(userID, candidate.UserID) || state.IsMutual() {
				continue
			}
		}
		result = append(result, candidate)
	}
	// Retain the existing deep-copy and non-nil result guarantees.
	return matching.FilterRestrictedCandidates(userID, result, nil, nil), nil
}
