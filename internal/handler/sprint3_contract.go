package handler

import (
	"context"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
)

// Discovery matches usecase.PairDiscoveryService. HTTP filtering and paging
// belong to the discovery handler, after Discover has applied LR2 rules.
type Discovery interface {
	Discover(ctx context.Context, userID int) ([]domain.Profile, error)
}

// Reactions matches usecase.ReactService. The current user comes from
// X-Demo-User middleware, never from a body or query parameter.
type Reactions interface {
	React(ctx context.Context, fromUserID, toUserID int) (interaction.State, error)
}

// Pairs matches interaction.Reader. Load(found=false) is an empty pair.
type Pairs interface {
	Load(pair interaction.Pair) (interaction.State, bool, error)
}

// These are wire DTOs, deliberately separate from domain.Profile and
// interaction.State. See docs/sprint-3-api-contract.md for their JSON meaning.
type DiscoveryItem struct {
	UserID    int      `json:"user_id"`
	City      string   `json:"city"`
	Interests []string `json:"interests"`
}

type DiscoveryResponse struct {
	Items  []DiscoveryItem `json:"items"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type PairResponse struct {
	TargetID int  `json:"target_id"`
	Liked    bool `json:"liked"`
	LikedBy  bool `json:"liked_by"`
	Mutual   bool `json:"mutual"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
