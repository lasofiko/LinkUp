package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

func main() {
	ctx := context.Background()

	users := memory.NewUserRepository()
	userService := usecase.NewUserUseCase(users)

	user := domain.User{
		ID:    1,
		Name:  "Sofa",
		Email: "sofa@example.com",
	}

	if err := userService.CreateUser(user); err != nil {
		log.Fatal(err)
	}

	savedUser, err := userService.GetUser(1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("LinkUp demo")
	fmt.Printf(
		"User: ID=%d | Name=%s | Email=%s\n",
		savedUser.ID,
		savedUser.Name,
		savedUser.Email,
	)

	profiles := memory.NewProfiles([]domain.Profile{
		{
			UserID:    1,
			City:      "Moscow",
			Available: true,
			Interests: []domain.Interest{"Go", "Music"},
		},
		{
			UserID:    2,
			City:      "Moscow",
			Available: true,
			Interests: []domain.Interest{"Go"},
		},
		{
			UserID:    3,
			City:      "Moscow",
			Available: false,
		},
		{
			UserID:    4,
			City:      "Moscow",
			Available: true,
		},
		{
			UserID:    5,
			City:      "Moscow",
			Available: true,
		},
	})

	likes := memory.NewLikes()
	matches := memory.NewMatches()

	likes.Add(1, 4)

	if err := matches.Save(ctx, 1, 5); err != nil {
		log.Fatal(err)
	}

	discovery := usecase.NewRestrictedDiscoveryService(
		profiles,
		likes,
		matches,
	)

	result, err := discovery.Discover(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}

	ids := make([]int, 0, len(result))
	for _, profile := range result {
		ids = append(ids, profile.UserID)
	}

	fmt.Println()
	fmt.Println("Contract B: restricted discovery")
	fmt.Println("Expected candidate IDs: [2]")
	fmt.Printf("Actual candidate IDs: %v\n", ids)
}