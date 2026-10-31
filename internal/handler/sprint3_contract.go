package handler

import (
	"context"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
)


type Discovery interface {
	Discover(ctx context.Context, userID int) ([]domain.Profile, error)
}


type Reactions interface {
	React(ctx context.Context, fromUserID, toUserID int) (interaction.State, error)
}


type Pairs interface {
	Load(pair interaction.Pair) (interaction.State, bool, error)
}

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
