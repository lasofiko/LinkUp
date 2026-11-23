package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type profileStoreStub struct {
	current domain.Profile
	updates int
}

func (s *profileStoreStub) GetProfile(_ context.Context, userID int) (domain.Profile, error) {
	if s.current.UserID != userID {
		return domain.Profile{}, domain.ErrProfileNotFound
	}
	return s.current, nil
}

func (s *profileStoreStub) UpdateProfile(_ context.Context, userID int, expectedVersion int64, city string, interests []domain.Interest) (domain.Profile, error) {
	s.updates++
	if s.current.UserID != userID {
		return domain.Profile{}, domain.ErrProfileNotFound
	}
	if s.current.Version != expectedVersion {
		return domain.Profile{}, domain.ErrVersionConflict
	}
	s.current.City = city
	s.current.Interests = append([]domain.Interest{}, interests...)
	s.current.Version++
	return s.current, nil
}

func initialProfile() *profileStoreStub {
	return &profileStoreStub{current: domain.Profile{
		UserID: 2, City: "Moscow", Interests: []domain.Interest{"Go"}, Available: true, Version: 1,
	}}
}

func TestProfileGetAndUpdate(t *testing.T) {
	store := initialProfile()
	service := NewProfileService(store)
	ctx := context.Background()

	initial, err := service.GetProfile(ctx, 2)
	if err != nil || initial.Version != 1 {
		t.Fatalf("get: profile=%+v err=%v", initial, err)
	}
	changed, err := service.UpdateProfile(ctx, 2, 2, UpdateProfileInput{
		ExpectedVersion: 1, City: " Paris ", Interests: []string{" Go ", "Music"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if changed.City != "Paris" || changed.Version != 2 || !changed.Available ||
		!reflect.DeepEqual(changed.Interests, []domain.Interest{"Go", "Music"}) || store.updates != 1 {
		t.Fatalf("unexpected update: %+v; calls=%d", changed, store.updates)
	}
}

func TestProfileInvalidEditDoesNotWrite(t *testing.T) {
	for _, tt := range []struct {
		name    string
		actorID int
		input   UpdateProfileInput
		want    error
	}{
		{"wrong actor", 3, UpdateProfileInput{1, "Paris", []string{"Go"}}, domain.ErrProfileForbidden},
		{"zero version", 2, UpdateProfileInput{0, "Paris", []string{"Go"}}, domain.ErrInvalidProfileEdit},
		{"blank city", 2, UpdateProfileInput{1, " ", []string{"Go"}}, domain.ErrInvalidProfileEdit},
		{"missing interests", 2, UpdateProfileInput{1, "Paris", nil}, domain.ErrInvalidProfileEdit},
		{"empty interest", 2, UpdateProfileInput{1, "Paris", []string{" "}}, domain.ErrInvalidProfileEdit},
		{"duplicate interest", 2, UpdateProfileInput{1, "Paris", []string{"Go", " go "}}, domain.ErrInvalidProfileEdit},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := initialProfile()
			_, err := NewProfileService(store).UpdateProfile(context.Background(), tt.actorID, 2, tt.input)
			if !errors.Is(err, tt.want) || store.updates != 0 {
				t.Fatalf("error=%v; updates=%d", err, store.updates)
			}
		})
	}
}

func TestProfileVersionConflictPreservesFirstEdit(t *testing.T) {
	store := initialProfile()
	service := NewProfileService(store)
	ctx := context.Background()
	_, err := service.UpdateProfile(ctx, 2, 2, UpdateProfileInput{1, "Paris", []string{"Music"}})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.UpdateProfile(ctx, 2, 2, UpdateProfileInput{1, "Berlin", []string{"Go"}})
	if !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("expected conflict, got: %v", err)
	}
	if store.current.City != "Paris" || store.current.Version != 2 ||
		!reflect.DeepEqual(store.current.Interests, []domain.Interest{"Music"}) {
		t.Fatalf("first edit was overwritten: %+v", store.current)
	}
}

func TestProfileNotFound(t *testing.T) {
	store := initialProfile()
	service := NewProfileService(store)
	_, err := service.GetProfile(context.Background(), 99)
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Fatalf("get error: %v", err)
	}
	_, err = service.UpdateProfile(context.Background(), 99, 99, UpdateProfileInput{1, "Paris", []string{}})
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Fatalf("update error: %v", err)
	}
}

func TestProfileContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := initialProfile()
	_, err := NewProfileService(store).UpdateProfile(ctx, 2, 2, UpdateProfileInput{1, "Paris", []string{}})
	if !errors.Is(err, context.Canceled) || store.updates != 0 {
		t.Fatalf("error=%v, updates=%d", err, store.updates)
	}
}
