package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type userReaderStub struct {
	users map[int]domain.User
	errs  map[int]error
}

func (s userReaderStub) GetByID(id int) (domain.User, error) {
	if err := s.errs[id]; err != nil {
		return domain.User{}, err
	}

	user, ok := s.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}

	return user, nil
}

type profileReaderStub struct {
	profiles map[int]domain.Profile
	err      error
}

func (s profileReaderStub) GetProfile(
	ctx context.Context,
	userID int,
) (domain.Profile, error) {
	if s.err != nil {
		return domain.Profile{}, s.err
	}

	profile, ok := s.profiles[userID]
	if !ok {
		return domain.Profile{}, domain.ErrProfileNotFound
	}

	return profile, nil
}

type likeReaderStub struct {
	exists bool
	err    error
}

func (s likeReaderStub) ExistsLike(
	ctx context.Context,
	from, to int,
) (bool, error) {
	if s.err != nil {
		return false, s.err
	}

	return s.exists, nil
}

func TestCandidateEligibilityServiceAllowed(t *testing.T) {
	users := userReaderStub{
		users: map[int]domain.User{
			1: {ID: 1},
			2: {ID: 2},
		},
		errs: map[int]error{},
	}

	profiles := profileReaderStub{
		profiles: map[int]domain.Profile{
			2: {
				UserID:    2,
				Available: true,
			},
		},
	}

	likes := likeReaderStub{
		exists: false,
	}

	service := NewCandidateEligibilityService(
		users,
		profiles,
		likes,
	)

	got, err := service.Check(context.Background(), 1, 2)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != domain.EligibilityAllowed {
		t.Fatalf("got %q, want %q", got, domain.EligibilityAllowed)
	}
}

func TestCandidateEligibilityServiceAlreadyLiked(t *testing.T) {
	users := userReaderStub{
		users: map[int]domain.User{
			1: {ID: 1},
			2: {ID: 2},
		},
		errs: map[int]error{},
	}

	profiles := profileReaderStub{
		profiles: map[int]domain.Profile{
			2: {
				UserID:    2,
				Available: true,
			},
		},
	}

	likes := likeReaderStub{
		exists: true,
	}

	service := NewCandidateEligibilityService(
		users,
		profiles,
		likes,
	)

	got, err := service.Check(context.Background(), 1, 2)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != domain.EligibilityAlreadyLiked {
		t.Fatalf(
			"got %q, want %q",
			got,
			domain.EligibilityAlreadyLiked,
		)
	}
}

func TestCandidateEligibilityServiceUnavailable(t *testing.T) {
	users := userReaderStub{
		users: map[int]domain.User{
			1: {ID: 1},
			2: {ID: 2},
		},
		errs: map[int]error{},
	}

	profiles := profileReaderStub{
		profiles: map[int]domain.Profile{
			2: {
				UserID:    2,
				Available: false,
			},
		},
	}

	service := NewCandidateEligibilityService(
		users,
		profiles,
		likeReaderStub{},
	)

	got, err := service.Check(context.Background(), 1, 2)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != domain.EligibilityUnavailable {
		t.Fatalf(
			"got %q, want %q",
			got,
			domain.EligibilityUnavailable,
		)
	}
}

func TestCandidateEligibilityServiceUserReaderError(t *testing.T) {
	errStorage := errors.New("storage unavailable")

	users := userReaderStub{
		users: map[int]domain.User{},
		errs: map[int]error{
			1: errStorage,
		},
	}

	service := NewCandidateEligibilityService(
		users,
		profileReaderStub{},
		likeReaderStub{},
	)

	_, err := service.Check(context.Background(), 1, 2)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errStorage) {
		t.Fatalf("lost original error: %v", err)
	}
}

func TestCandidateEligibilityServiceCandidateReaderError(t *testing.T) {
	errStorage := errors.New("candidate storage error")

	users := userReaderStub{
		users: map[int]domain.User{
			1: {ID: 1},
		},
		errs: map[int]error{
			2: errStorage,
		},
	}

	service := NewCandidateEligibilityService(
		users,
		profileReaderStub{},
		likeReaderStub{},
	)

	_, err := service.Check(context.Background(), 1, 2)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errStorage) {
		t.Fatalf("lost original error: %v", err)
	}
}

func TestCandidateEligibilityServiceProfileReaderError(t *testing.T) {
	errStorage := errors.New("profile storage error")

	users := userReaderStub{
		users: map[int]domain.User{
			1: {ID: 1},
			2: {ID: 2},
		},
		errs: map[int]error{},
	}

	profiles := profileReaderStub{
		err: errStorage,
	}

	service := NewCandidateEligibilityService(
		users,
		profiles,
		likeReaderStub{},
	)

	_, err := service.Check(context.Background(), 1, 2)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errStorage) {
		t.Fatalf("lost original error: %v", err)
	}
}

func TestCandidateEligibilityServiceLikeReaderError(t *testing.T) {
	errStorage := errors.New("like storage error")

	users := userReaderStub{
		users: map[int]domain.User{
			1: {ID: 1},
			2: {ID: 2},
		},
		errs: map[int]error{},
	}

	profiles := profileReaderStub{
		profiles: map[int]domain.Profile{
			2: {
				UserID:    2,
				Available: true,
			},
		},
	}

	likes := likeReaderStub{
		err: errStorage,
	}

	service := NewCandidateEligibilityService(
		users,
		profiles,
		likes,
	)

	_, err := service.Check(context.Background(), 1, 2)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errStorage) {
		t.Fatalf("lost original error: %v", err)
	}
}

func TestCandidateEligibilityServiceUserNotFound(t *testing.T) {
	users := userReaderStub{
		users: map[int]domain.User{},
		errs:  map[int]error{},
	}

	service := NewCandidateEligibilityService(
		users,
		profileReaderStub{},
		likeReaderStub{},
	)

	_, err := service.Check(context.Background(), 1, 2)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCandidateEligibilityServiceCandidateNotFound(t *testing.T) {
	users := userReaderStub{
		users: map[int]domain.User{
			1: {ID: 1},
		},
		errs: map[int]error{},
	}

	service := NewCandidateEligibilityService(
		users,
		profileReaderStub{},
		likeReaderStub{},
	)

	_, err := service.Check(context.Background(), 1, 2)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCandidateEligibilityServiceSelfInteraction(t *testing.T) {
	users := userReaderStub{
		users: map[int]domain.User{
			1: {ID: 1},
		},
		errs: map[int]error{},
	}

	profiles := profileReaderStub{
		profiles: map[int]domain.Profile{
			1: {
				UserID:    1,
				Available: true,
			},
		},
	}

	service := NewCandidateEligibilityService(
		users,
		profiles,
		likeReaderStub{},
	)

	got, err := service.Check(context.Background(), 1, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != domain.EligibilityUnavailable {
		t.Fatalf(
			"got %q, want %q",
			got,
			domain.EligibilityUnavailable,
		)
	}
}
