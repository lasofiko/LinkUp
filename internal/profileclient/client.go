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

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func (c *Client) GetProfile(context.Context, int) (domain.Profile, error) {
	return domain.Profile{}, ErrNotImplemented
}
