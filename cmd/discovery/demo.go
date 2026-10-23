package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/matching"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

func main() {
	ctx := context.Background()

	store := interaction.NewMemory()

	userRepo := memory.NewUserRepository()
	for i := 1; i <= 6; i++ {
		err := userRepo.Create(domain.User{
			ID:   i,
			Name: fmt.Sprintf("User %d", i),
		})
		if err != nil {
			log.Fatalf("failed to create user %d: %v", i, err)
		}
	}

	profiles := []domain.Profile{
		{UserID: 1, City: "Москва", Interests: interests("go", "listening to music", "videogames"), Available: true},
		{UserID: 2, City: "Москва", Interests: interests("go", "videogames", "playing chess"), Available: true},
		{UserID: 3, City: "Москва", Interests: interests("listening to music", "cooking"), Available: true},
		{UserID: 4, City: "Санкт-Петербург", Interests: interests("go", "listening to music"), Available: true},
		{UserID: 5, City: "Москва", Interests: interests("go", "listening to music"), Available: false},
		{UserID: 6, City: "Москва", Interests: interests("playing chess", "cooking"), Available: true},
	}
	profileRepo := memory.NewProfiles(profiles)

	baseDiscovery := matching.NewDiscoveryService(profileRepo)
	discoverySvc := usecase.NewPairDiscoveryService(baseDiscovery, store)

	reactSvc := usecase.NewReactService(userRepo, profileRepo, store)

	userA := 1
	userB := 2

	fmt.Println("1. Starting state")
	printDiscovery(ctx, discoverySvc, userA)
	printDiscovery(ctx, discoverySvc, userB)

	fmt.Println("\n2. User 1 likes User 2")
	state, err := reactSvc.React(ctx, userA, userB)
	if err != nil {
		log.Fatalf("React failed: %v", err)
	}
	fmt.Printf("Pair state: %+v\n", state)

	fmt.Println("\n3. state after User 1 likes User 2")
	printDiscovery(ctx, discoverySvc, userA)
	printDiscovery(ctx, discoverySvc, userB)

	fmt.Println("\n4. User 2 returns the like to User 1, match is formed")
	state, err = reactSvc.React(ctx, userB, userA)
	if err != nil {
		log.Fatalf("React failed: %v", err)
	}
	fmt.Printf("Pair state (IsMutual=%v): %+v\n", state.IsMutual(), state)

	fmt.Println("\n5. State after the match")
	printDiscovery(ctx, discoverySvc, userA)
	printDiscovery(ctx, discoverySvc, userB)

	fmt.Println("\n6. Repeated like is attempted")
	state, err = reactSvc.React(ctx, userA, userB)
	fmt.Printf("Result of the repeat: err=%v, state=%+v\n", err, state)
}

func printDiscovery(ctx context.Context, svc *usecase.PairDiscoveryService, userID int) {
	result, err := svc.Discover(ctx, userID)
	if err != nil {
		log.Printf("error discovering for user %d: %v", userID, err)
		return
	}
	fmt.Printf("Result for user %d (%d candidates):\n", userID, len(result))
	for i, p := range result {
		fmt.Printf("  %d. user %d, %s\n", i+1, p.UserID, p.City)
	}
}

func interests(values ...string) []domain.Interest {
	out := make([]domain.Interest, 0, len(values))
	for _, v := range values {
		interest, err := domain.NewInterest(v)
		if err != nil {
			log.Fatalf("invalid interest %q: %v", v, err)
		}
		out = append(out, interest)
	}
	return out
}
