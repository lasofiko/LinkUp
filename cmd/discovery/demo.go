package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/matching"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
)

func main() {
	profiles := []domain.Profile{
		{UserID: 1, City: "Москва", Interests: interests("go", "listening to music", "videogames"), Available: true},
		{UserID: 2, City: "Москва", Interests: interests("go", "videogames", "playing chess"), Available: true},
		{UserID: 3, City: "Москва", Interests: interests("listening to music", "cooking"), Available: true},
		{UserID: 4, City: "Санкт-Петербург", Interests: interests("go", "listening to music"), Available: true},
		{UserID: 5, City: "Москва", Interests: interests("go", "listening to music"), Available: false},
		{UserID: 6, City: "Москва", Interests: interests("playing chess", "cooking"), Available: true},
	}

	repo := memory.NewProfiles(profiles)
	service := matching.NewDiscoveryService(repo)

	const targetID = 1

	result, err := service.Recommend(context.Background(), targetID)
	if err != nil {
		log.Fatalf("recommend for user %d failed: %v", targetID, err)
	}

	fmt.Printf("suggestions for user %d:\n", targetID)
	if len(result) == 0 {
		fmt.Println("not found")
		return
	}
	for i, p := range result {
		fmt.Printf("%d. user %d, %s, interests: %v\n", i+1, p.UserID, p.City, p.Interests)
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
