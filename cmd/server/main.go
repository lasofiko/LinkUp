package main

import (
	"fmt"
	"log"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

func main() {
	userRepository := memory.NewUserRepository()

	userUseCase := usecase.NewUserUseCase(userRepository)

	user := domain.User{
		ID:        1,
		Name:      "Sofa",
		Email:     "sofa@example.com",
		City:      "Moscow",
		Interests: []string{"Go", "Music", "Movies"},
		Available: true,
	}

	err := userUseCase.CreateUser(user)
	if err != nil {
		log.Fatal(err)
	}

	savedUser, err := userUseCase.GetUser(1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("LinkUp demo")
	fmt.Println("User successfully created")
	fmt.Printf(
		"ID: %d | Name: %s | Email: %s | City: %s\n",
		savedUser.ID,
		savedUser.Name,
		savedUser.Email,
		savedUser.City,
	)
	candidates := []domain.User{
		{
			ID:        2,
			Name:      "Available candidate",
			Available: true,
			Interests: []string{"Go"},
		},
		{
			ID:        3,
			Name:      "Unavailable candidate",
			Available: false,
		},
		{
			ID:        4,
			Name:      "Already liked candidate",
			Available: true,
		},
		{
			ID:        5,
			Name:      "Matched candidate",
			Available: true,
		},
	}

	for _, candidate := range candidates {
		if err := userUseCase.CreateUser(candidate); err != nil {
			log.Fatal(err)
		}
	}

	restrictions := memory.NewDiscoveryRepository()
	restrictions.AddLike(1, 4)
	restrictions.AddMatch(1, 5)

	discovery := usecase.NewRestrictedDiscoveryService(
		userRepository,
		restrictions,
		restrictions,
	)

	result, err := discovery.Discover(1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("Contract B: restricted discovery")
	fmt.Println("Expected candidate IDs: [2]")
	fmt.Printf("Actual candidates: %+v\n", result)
}
