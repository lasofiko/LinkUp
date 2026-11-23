package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

const maxProfileEditBodyBytes int64 = 64 << 10

type ProfileService interface {
	GetProfile(ctx context.Context, userID int) (domain.Profile, error)
	UpdateProfile(ctx context.Context, actorID, userID int, input usecase.UpdateProfileInput) (domain.Profile, error)
}

type ProfileResponse struct {
	UserID    int      `json:"user_id"`
	City      string   `json:"city"`
	Interests []string `json:"interests"`
	Available bool     `json:"available"`
	Version   int64    `json:"version"`
}

type UpdateProfileRequest struct {
	ExpectedVersion int64    `json:"expected_version"`
	City            string   `json:"city"`
	Interests       []string `json:"interests"`
}

func profileResponse(p domain.Profile) ProfileResponse {
	interests := make([]string, len(p.Interests))
	for i, interest := range p.Interests {
		interests[i] = string(interest)
	}
	return ProfileResponse{
		UserID:    p.UserID,
		City:      p.City,
		Interests: interests,
		Available: p.Available,
		Version:   p.Version,
	}
}

func profileRequestIDs(r *http.Request) (actorID, userID int, ok bool) {
	if r.URL.RawQuery != "" {
		return 0, 0, false
	}
	actorID, ok = r.Context().Value(currentUserIDKey{}).(int)
	if !ok || actorID <= 0 {
		return 0, 0, false
	}
	userID, ok = positiveID(r.PathValue("userID"))
	return actorID, userID, ok
}

func writeProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrProfileForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "cannot edit another user's profile")
	case errors.Is(err, domain.ErrProfileNotFound), errors.Is(err, domain.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "not_found", "profile not found")
	case errors.Is(err, domain.ErrVersionConflict):
		writeError(w, http.StatusConflict, "version_conflict", "profile has a newer version")
	case errors.Is(err, domain.ErrInvalidProfileEdit):
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid profile edit")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func GetProfileHandler(service ProfileService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, userID, ok := profileRequestIDs(r)
		if !ok || !emptyBody(r) {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid request")
			return
		}
		profile, err := service.GetProfile(r.Context(), userID)
		if err != nil {
			writeProfileError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, profileResponse(profile))
	})
}

func UpdateProfileHandler(service ProfileService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, userID, ok := profileRequestIDs(r)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid request")
			return
		}

		if actorID != userID {
			writeError(w, http.StatusForbidden, "forbidden", "cannot edit another user's profile")
			return
		}

		if r.Body == nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "missing JSON body")
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, maxProfileEditBodyBytes+1))
		if err != nil || int64(len(body)) > maxProfileEditBodyBytes {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid or oversized JSON body")
			return
		}

		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		var input UpdateProfileRequest
		if err := decoder.Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
			return
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "invalid_request", "trailing JSON content")
			return
		}

		profile, err := service.UpdateProfile(r.Context(), actorID, userID, usecase.UpdateProfileInput{
			ExpectedVersion: input.ExpectedVersion,
			City:            input.City,
			Interests:       input.Interests,
		})
		if err != nil {
			writeProfileError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, profileResponse(profile))
	})
}
