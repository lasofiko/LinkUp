package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/lasofiko/LinkUp/internal/interaction"
)

var ErrPairBlocked = errors.New("pair is blocked")

// BlockStore must save the block and deactivate the match atomically.
// Its PostgreSQL implementation is provided by the database task.
type BlockStore interface {
	Block(ctx context.Context, pair interaction.Pair) (interaction.State, error)
}

type BlockService struct {
	users EligibilityUserReader
	store BlockStore
}

func NewBlockService(users EligibilityUserReader, store BlockStore) *BlockService {
	return &BlockService{users: users, store: store}
}

func (s *BlockService) Block(ctx context.Context, actorID, targetID int) (interaction.State, error) {
	if err := ctx.Err(); err != nil {
		return interaction.State{}, err
	}
	if actorID <= 0 || targetID <= 0 {
		return interaction.State{}, ErrInvalidBlockID
	}
	if actorID == targetID {
		return interaction.State{}, interaction.ErrSameUser
	}
	if _, err := s.users.GetByID(actorID); err != nil {
		return interaction.State{}, fmt.Errorf("get actor: %w", err)
	}
	if _, err := s.users.GetByID(targetID); err != nil {
		return interaction.State{}, fmt.Errorf("get target: %w", err)
	}
	pair, err := interaction.NewPair(actorID, targetID)
	if err != nil {
		return interaction.State{}, err
	}
	state, err := s.store.Block(ctx, pair)
	if err != nil {
		return interaction.State{}, fmt.Errorf("block pair: %w", err)
	}
	return state, nil
}

var ErrInvalidBlockID = errors.New("invalid block user id")
