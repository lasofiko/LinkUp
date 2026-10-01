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
		ID:       1,
		Name:     "Sofa",
		Email:    "sofa@example.com",
		City:     "Moscow",
		Interests: []string{"Go", "Music", "Movies"},
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
}