package memory

import (
	"context"
)

type FailingLikes struct {
	Err error
}

func (f FailingLikes) ExistsLike(_ context.Context, _, _ int) (bool, error) {
	return false, f.Err
}

type FailingMatches struct {
	ExistsErr error
	SaveErr   error
}

func (f FailingMatches) ExistsMatch(_ context.Context, _, _ int) (bool, error) {
	return false, f.ExistsErr
}

func (f FailingMatches) Save(_ context.Context, _, _ int) error {
	return f.SaveErr
}