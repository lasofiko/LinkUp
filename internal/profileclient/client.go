package profileclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/lasofiko/LinkUp/internal/domain"
)

const DefaultMaxBodyBytes int64 = 64 << 10

var (
	ErrDependency    = errors.New("profile API dependency failure")
	ErrTimeout       = errors.New("profile API timeout")
	ErrInvalidUserID = errors.New("invalid profile user ID")
)

type profileResponse struct {
	UserID    *int      `json:"user_id"`
	City      *string   `json:"city"`
	Interests *[]string `json:"interests"`
	Available *bool     `json:"available"`
}

type Client struct {
	BaseURL      string
	HTTP         *http.Client
	MaxBodyBytes int64
}

func NewClient(
	baseURL string,
	httpClient *http.Client,
	maxBodyBytes int64,
) (*Client, error) {
	if httpClient == nil {
		return nil, errors.New("HTTP client is nil")
	}

	if maxBodyBytes <= 0 || maxBodyBytes == math.MaxInt64 {
		return nil, errors.New("invalid response body limit")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	if (parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Host == "" ||
		parsed.User != nil ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return nil, errors.New("invalid profile API base URL")
	}

	return &Client{
		BaseURL:      baseURL,
		HTTP:         httpClient,
		MaxBodyBytes: maxBodyBytes,
	}, nil
}

func classifyRequestError(ctx context.Context, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) ||
		errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %w", ErrTimeout, err)
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("%w: %w", ErrTimeout, err)
	}

	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("request canceled: %w", err)
	}

	if errors.Is(ctx.Err(), context.Canceled) {
		return fmt.Errorf("request canceled: %w: %w", ctx.Err(), err)
	}

	return fmt.Errorf("%w: %w", ErrDependency, err)
}

func (c *Client) GetProfile(
	ctx context.Context,
	userID int,
) (domain.Profile, error) {
	if userID <= 0 {
		return domain.Profile{}, ErrInvalidUserID
	}

	endpoint := strings.TrimRight(c.BaseURL, "/") +
		"/v1/profiles/" + url.PathEscape(strconv.Itoa(userID))

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return domain.Profile{}, fmt.Errorf(
			"%w: create request: %w", ErrDependency, err,
		)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return domain.Profile{}, classifyRequestError(ctx, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Profile{}, fmt.Errorf(
			"profile %d: %w", userID, domain.ErrProfileNotFound,
		)
	}

	if resp.StatusCode != http.StatusOK {
		return domain.Profile{}, fmt.Errorf(
			"%w: HTTP status %d", ErrDependency, resp.StatusCode,
		)
	}

	body, err := io.ReadAll(
		io.LimitReader(resp.Body, c.MaxBodyBytes+1),
	)
	if err != nil {
		return domain.Profile{}, classifyRequestError(ctx, err)
	}

	if int64(len(body)) > c.MaxBodyBytes {
		return domain.Profile{}, fmt.Errorf(
			"%w: response body exceeds limit", ErrDependency,
		)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	var data profileResponse
	if err := decoder.Decode(&data); err != nil {
		return domain.Profile{}, fmt.Errorf(
			"%w: invalid JSON: %w", ErrDependency, err,
		)
	}

	var extra any

	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return domain.Profile{}, fmt.Errorf(
				"%w: invalid trailing JSON: %w", ErrDependency, err,
			)
		}

		return domain.Profile{}, fmt.Errorf(
			"%w: extra JSON value", ErrDependency,
		)
	}

	if data.UserID == nil ||
		data.City == nil ||
		data.Interests == nil ||
		data.Available == nil {
		return domain.Profile{}, fmt.Errorf(
			"%w: missing required fields", ErrDependency,
		)
	}

	if *data.UserID != userID || strings.TrimSpace(*data.City) == "" {
		return domain.Profile{}, fmt.Errorf(
			"%w: invalid profile data", ErrDependency,
		)
	}

	interests := make([]domain.Interest, 0, len(*data.Interests))

	for _, value := range *data.Interests {
		interest, err := domain.NewInterest(value)
		if err != nil {
			return domain.Profile{}, fmt.Errorf(
				"%w: invalid interest: %w", ErrDependency, err,
			)
		}

		interests = append(interests, interest)
	}

	return domain.Profile{
		UserID:    *data.UserID,
		City:      *data.City,
		Interests: interests,
		Available: *data.Available,
	}, nil
}
