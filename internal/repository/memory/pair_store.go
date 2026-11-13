package memory

import (
	"errors"

	"github.com/lasofiko/LinkUp/internal/interaction"
)

var ErrPairStoreNotImplemented = errors.New("memory pair store is not implemented")

type PairStore struct{}

func NewPairStore() *PairStore { return &PairStore{} }

func (*PairStore) Load(interaction.Pair) (interaction.State, bool, error) {
	return interaction.State{}, false, ErrPairStoreNotImplemented
}

func (*PairStore) Save(interaction.State) error { return ErrPairStoreNotImplemented }

var _ interaction.Store = (*PairStore)(nil)
