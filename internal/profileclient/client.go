package profileclient

import (
	"context"
	"errors"
	"net/http"

	"github.com/lasofiko/LinkUp/internal/domain"
)

var (
	ErrNotImplemented = errors.New("profile API client is not implemented")
	ErrDependency     = errors.New("profile API dependency failure")
	ErrTimeout        = errors.New("profile API timeout")
)

// Client is intended to satisfy usecase.EligibilityProfileReader and the
// GetProfile half of matching.ProfileRepository. Discovery still needs a
// ListProfiles source; see the API contract's split-source decision.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// TODO(Оля): use context, bounded response reads, JSON validation and
// dependency error types preserving errors.Is/errors.As.
func (c *Client) GetProfile(context.Context, int) (domain.Profile, error) {
	return domain.Profile{}, ErrNotImplemented
}
