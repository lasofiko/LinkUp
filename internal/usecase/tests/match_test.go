package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lasofiko/LinkUp/internal/matching"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

func TestMatchService_NoLikes(t *testing.T) {
	svc := usecase.NewMatchService(memory.NewLikes(), memory.NewMatches())

	got, err := svc.GetResult(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != matching.NoMatch {
		t.Fatalf("want NoMatch, got %v", got)
	}
}

func TestMatchService_OneWayLike(t *testing.T) {
	likes := memory.NewLikes()
	matches := memory.NewMatches()
	likes.Add(1, 2)

	svc := usecase.NewMatchService(likes, matches)
	got, err := svc.GetResult(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != matching.NoMatch {
		t.Fatalf("want NoMatch, got %v", got)
	}
	if matches.Count() != 0 {
		t.Fatalf("Save must not be called, matches=%d", matches.Count())
	}
}

func TestMatchService_CreatesMatch(t *testing.T) {
	likes := memory.NewLikes()
	matches := memory.NewMatches()
	likes.Add(1, 2)
	likes.Add(2, 1)

	svc := usecase.NewMatchService(likes, matches)
	got, err := svc.GetResult(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != matching.MatchCreated {
		t.Fatalf("want MatchCreated, got %v", got)
	}
	if matches.Count() != 1 {
		t.Fatalf("want exactly 1 match, got %d", matches.Count())
	}
}

func TestMatchService_ReturnsExistingMatch(t *testing.T) {
	likes := memory.NewLikes()
	matches := memory.NewMatches()
	likes.Add(1, 2)
	likes.Add(2, 1)

	svc := usecase.NewMatchService(likes, matches)

	if _, err := svc.GetResult(context.Background(), 1, 2); err != nil {
		t.Fatalf("first call: %v", err)
	}
	got, err := svc.GetResult(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if got != matching.ExistingMatch {
		t.Fatalf("want ExistingMatch, got %v", got)
	}
	if matches.Count() != 1 {
		t.Fatalf("must not duplicate match, got %d", matches.Count())
	}
}

func TestMatchService_DoesNotCreateDuplicateForReversedUsers(t *testing.T) {
	likes := memory.NewLikes()
	matches := memory.NewMatches()
	likes.Add(1, 2)
	likes.Add(2, 1)

	svc := usecase.NewMatchService(likes, matches)

	if _, err := svc.GetResult(context.Background(), 1, 2); err != nil {
		t.Fatalf("(1,2): %v", err)
	}
	got, err := svc.GetResult(context.Background(), 2, 1)
	if err != nil {
		t.Fatalf("(2,1): %v", err)
	}
	if got != matching.ExistingMatch {
		t.Fatalf("want ExistingMatch, got %v", got)
	}
	if matches.Count() != 1 {
		t.Fatalf("A-B and B-A must be one match, got %d", matches.Count())
	}
}

func TestMatchService_LikesReadError(t *testing.T) {
	sentinel := errors.New("likes db down")
	svc := usecase.NewMatchService(
		memory.FailingLikes{Err: sentinel},
		memory.NewMatches(),
	)

	_, err := svc.GetResult(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, usecase.ErrReadLikes) {
		t.Fatalf("want ErrReadLikes, got %v", err)
	}
}

func TestMatchService_MatchReadError(t *testing.T) {
	likes := memory.NewLikes()
	likes.Add(1, 2)
	likes.Add(2, 1)

	sentinel := errors.New("matches db down")
	svc := usecase.NewMatchService(
		likes,
		memory.FailingMatches{ExistsErr: sentinel},
	)

	_, err := svc.GetResult(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, usecase.ErrReadMatch) {
		t.Fatalf("want ErrReadMatch, got %v", err)
	}
}

func TestMatchService_MatchSaveError(t *testing.T) {
	likes := memory.NewLikes()
	likes.Add(1, 2)
	likes.Add(2, 1)

	sentinel := errors.New("matches write failed")
	svc := usecase.NewMatchService(
		likes,
		memory.FailingMatches{SaveErr: sentinel},
	)

	_, err := svc.GetResult(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, usecase.ErrSaveMatch) {
		t.Fatalf("want ErrSaveMatch, got %v", err)
	}
}