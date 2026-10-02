package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/lasofiko/LinkUp/internal/matching"
)

var (
	ErrReadLikes = errors.New("read likes failed")
	ErrReadMatch = errors.New("read match failed")
	ErrSaveMatch = errors.New("save match failed")
)

type LikeRepository interface {
	ExistsLike(ctx context.Context, from, to int) (bool, error)
}

type MatchRepository interface {
	ExistsMatch(ctx context.Context, a, b int) (bool, error)
	Save(ctx context.Context, a, b int) error
}

type MatchService struct {
	likes   LikeRepository
	matches MatchRepository
}

func NewMatchService(likes LikeRepository, matches MatchRepository) *MatchService {
	return &MatchService{likes: likes, matches: matches}
}

func (s *MatchService) GetResult(ctx context.Context, userA, userB int) (matching.MatchResult, error) {
	likeAB, err := s.likes.ExistsLike(ctx, userA, userB)
	if err != nil {
		return matching.NoMatch, fmt.Errorf("%w: %d->%d: %v", ErrReadLikes, userA, userB, err)
	}
	likeBA, err := s.likes.ExistsLike(ctx, userB, userA)
	if err != nil {
		return matching.NoMatch, fmt.Errorf("%w: %d->%d: %v", ErrReadLikes, userB, userA, err)
	}
	matchExists, err := s.matches.ExistsMatch(ctx, userA, userB)
	if err != nil {
		return matching.NoMatch, fmt.Errorf("%w: %d-%d: %v", ErrReadMatch, userA, userB, err)
	}

	result := matching.DetermineMatch(likeAB, likeBA, matchExists)

	if result == matching.MatchCreated {
		if err := s.matches.Save(ctx, userA, userB); err != nil {
			return matching.NoMatch, fmt.Errorf("%w: %d-%d: %v", ErrSaveMatch, userA, userB, err)
		}
	}
	return result, nil
}