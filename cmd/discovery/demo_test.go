package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/matching"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

func TestIntegration_DiscoveryAndReactLifecycle(t *testing.T) {
	ctx := context.Background()

	store := interaction.NewMemory()

	userRepo := memory.NewUserRepository()
	for i := 1; i <= 6; i++ {
		if err := userRepo.Create(domain.User{ID: i, Name: fmt.Sprintf("User %d", i)}); err != nil {
			t.Fatalf("failed to setup user %d: %v", i, err)
		}
	}

	profiles := []domain.Profile{
		{UserID: 1, City: "Москва", Interests: mockInterests("go", "boardgames"), Available: true},
		{UserID: 2, City: "Москва", Interests: mockInterests("go", "boardgames"), Available: true},
		{UserID: 3, City: "Москва", Interests: mockInterests("cooking"), Available: true},
	}
	profileRepo := memory.NewProfiles(profiles)

	baseDiscovery := matching.NewDiscoveryService(profileRepo)
	discoverySvc := usecase.NewPairDiscoveryService(baseDiscovery, store)
	reactSvc := usecase.NewReactService(userRepo, profileRepo, store)

	userA, userB := 1, 2

	resA, _ := discoverySvc.Discover(ctx, userA)
	if len(resA) < 1 || resA[0].UserID != userB {
		t.Fatalf("expected user %d to see user %d initially", userA, userB)
	}

	state, err := reactSvc.React(ctx, userA, userB)
	if err != nil {
		t.Fatalf("unexpected react error: %v", err)
	}
	if !state.FirstLikesSecond {
		t.Fatal("expected FirstLikesSecond to be true")
	}

	resA, _ = discoverySvc.Discover(ctx, userA)
	resB, _ := discoverySvc.Discover(ctx, userB)

	for _, p := range resA {
		if p.UserID == userB {
			t.Fatalf("user %d should be hidden from user %d after like", userB, userA)
		}
	}
	if len(resB) < 1 || resB[0].UserID != userA {
		t.Fatalf("user %d should still see user %d before mutual reaction", userB, userA)
	}

	state, err = reactSvc.React(ctx, userB, userA)
	if err != nil {
		t.Fatalf("unexpected mutual react error: %v", err)
	}
	if !state.IsMutual() {
		t.Fatal("expected pair state to be mutual")
	}

	resA, _ = discoverySvc.Discover(ctx, userA)
	resB, _ = discoverySvc.Discover(ctx, userB)

	for _, p := range resA {
		if p.UserID == userB {
			t.Fatalf("user %d should be hidden from user %d after match", userB, userA)
		}
	}
	for _, p := range resB {
		if p.UserID == userA {
			t.Fatalf("user %d should be hidden from user %d after match", userA, userB)
		}
	}

	stateRepeat, err := reactSvc.React(ctx, userA, userB)
	if err != nil {
		t.Fatalf("repeat react should not fail, got: %v", err)
	}
	if stateRepeat != state {
		t.Fatalf("expected state to remain unchanged on repeat like")
	}
}

func mockInterests(values ...string) []domain.Interest {
	out := make([]domain.Interest, 0, len(values))
	for _, v := range values {
		interest, _ := domain.NewInterest(v)
		out = append(out, interest)
	}
	return out
}
