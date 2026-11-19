package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
)

var ErrReactionUnavailable = errors.New("reaction is unavailable")

type ReactService struct {
	users    EligibilityUserReader
	profiles EligibilityProfileReader
	store    interaction.Store
}

func NewReactService(users EligibilityUserReader, profiles EligibilityProfileReader, store interaction.Store) *ReactService {
	return &ReactService{users: users, profiles: profiles, store: store}
}

// React applies a directed like. A successful repeat returns the current state
// without saving. On every error the returned state is zero.
func (s *ReactService) React(ctx context.Context, fromUserID, toUserID int) (interaction.State, error) {
	if err := ctx.Err(); err != nil {
		return interaction.State{}, err
	}
	user, err := s.users.GetByID(fromUserID)
	if err != nil {
		return interaction.State{}, fmt.Errorf("get reacting user %d: %w", fromUserID, err)
	}
	candidate, err := s.users.GetByID(toUserID)
	if err != nil {
		return interaction.State{}, fmt.Errorf("get reaction candidate %d: %w", toUserID, err)
	}
	profile, err := s.profiles.GetProfile(ctx, candidate.ID)
	if err != nil {
		return interaction.State{}, fmt.Errorf("get reaction candidate profile %d: %w", candidate.ID, err)
	}
	pair, err := interaction.NewPair(user.ID, candidate.ID)
	if err != nil {
		return interaction.State{}, fmt.Errorf("create reaction pair: %w", err)
	}
	state, found, err := s.store.Load(pair)
	if err != nil {
		return interaction.State{}, fmt.Errorf("load reaction pair %v: %w", pair, err)
	}
	if !found {
		state = interaction.State{}
	} else if state.Pair != pair {
		return interaction.State{}, fmt.Errorf("load reaction pair %v: %w", pair, interaction.ErrStatePair)
	}
	status := domain.CheckCandidateEligibility(user.ID, profile, state.HasLike(user.ID, candidate.ID))
	if status == domain.EligibilityUnavailable {
		return interaction.State{}, fmt.Errorf("reaction %d -> %d: %w", user.ID, candidate.ID, ErrReactionUnavailable)
	}
	next, changed, err := interaction.ApplyLike(state, user.ID, candidate.ID)
	if err != nil {
		return interaction.State{}, fmt.Errorf("apply reaction: %w", err)
	}
	if changed {
		if err := ctx.Err(); err != nil {
			return interaction.State{}, err
		}
		if err := s.store.Save(next); err != nil {
			return interaction.State{}, fmt.Errorf("save reaction pair %v: %w", pair, err)
		}
	}
	return next, nil
}
