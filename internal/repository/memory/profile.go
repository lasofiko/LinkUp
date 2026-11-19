package memory

import (
	"context"
	"fmt"
	"slices"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type Profiles struct {
	items []domain.Profile
}

func NewProfiles(size []domain.Profile) *Profiles {
	items := make([]domain.Profile, len(size))
	for i, p := range size {
		items[i] = clone(p)
	}
	return &Profiles{items: items}
}

func (r *Profiles) GetProfile(ctx context.Context, userID int) (domain.Profile, error) {
	for _, p := range r.items {
		if p.UserID == userID {
			return clone(p), nil
		}
	}
	return domain.Profile{}, fmt.Errorf("profile of user %d: %w", userID, domain.ErrProfileNotFound)
}

func (r *Profiles) ListProfiles(ctx context.Context) ([]domain.Profile, error) {
	out := make([]domain.Profile, len(r.items))
	for i, p := range r.items {
		out[i] = clone(p)
	}
	return out, nil
}

func clone(p domain.Profile) domain.Profile {
	p.Interests = slices.Clone(p.Interests)
	return p
}
