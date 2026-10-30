package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/profileclient"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

type currentUserIDKey struct{}

// WithCurrentUserID is the handoff from identity middleware to business handlers.
func WithCurrentUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, currentUserIDKey{}, userID)
}

func requestIDs(r *http.Request) (int, int, bool) {
	if r.URL.RawQuery != "" {
		return 0, 0, false
	}
	userID, ok := r.Context().Value(currentUserIDKey{}).(int)
	if !ok || userID <= 0 {
		return 0, 0, false
	}
	target, ok := positiveID(r.PathValue("targetID"))
	return userID, target, ok
}

func positiveID(raw string) (int, bool) {
	if raw == "" || raw[0] == '0' {
		return 0, false
	}
	for _, c := range raw {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	id, err := strconv.Atoi(raw)
	return id, err == nil && id > 0
}

func emptyBody(r *http.Request) bool {
	if r.Body == nil {
		return true
	}
	var b [1]byte
	n, err := r.Body.Read(b[:])
	return n == 0 && (err == nil || errors.Is(err, io.EOF))
}

func pairDTO(userID, targetID int, state interaction.State) PairResponse {
	return PairResponse{
		TargetID: targetID,
		Liked:    state.HasLike(userID, targetID),
		LikedBy:  state.HasLike(targetID, userID),
		Mutual:   state.IsMutual(),
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Code: code, Message: message})
}

func writeHandlerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrReactionUnavailable):
		writeError(w, http.StatusForbidden, "reaction_unavailable", "reaction is unavailable")
	case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, memory.ErrUserNotFound), errors.Is(err, domain.ErrProfileNotFound):
		writeError(w, http.StatusNotFound, "not_found", "required user or profile was not found")
	case errors.Is(err, interaction.ErrSameUser):
		writeError(w, http.StatusConflict, "invalid_transition", "cannot interact with yourself")
	case errors.Is(err, profileclient.ErrTimeout):
		writeError(w, http.StatusGatewayTimeout, "dependency_timeout", "profile service timed out")
	case errors.Is(err, profileclient.ErrDependency):
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			writeError(w, http.StatusGatewayTimeout, "dependency_timeout", "profile service timed out")
			return
		}
		writeError(w, http.StatusBadGateway, "dependency_failure", "profile service failed")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

// LikeHandler applies a directed like through ReactService and returns its pair state.
func LikeHandler(reactions Reactions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, targetID, ok := requestIDs(r)
		if !ok || !emptyBody(r) {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid request")
			return
		}
		if userID == targetID {
			writeError(w, http.StatusConflict, "invalid_transition", "cannot interact with yourself")
			return
		}
		state, err := reactions.React(r.Context(), userID, targetID)
		if err != nil {
			writeHandlerError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, pairDTO(userID, targetID, state))
	})
}

// PairHandler reads the same store used by ReactService, without applying a like.
func PairHandler(pairs Pairs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, targetID, ok := requestIDs(r)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid request")
			return
		}
		pair, err := interaction.NewPair(userID, targetID)
		if err != nil {
			writeHandlerError(w, err)
			return
		}
		state, found, err := pairs.Load(pair)
		if err != nil {
			writeHandlerError(w, err)
			return
		}
		if !found {
			state = interaction.State{}
		} else if state.Pair != pair {
			writeHandlerError(w, interaction.ErrStatePair)
			return
		}
		writeJSON(w, http.StatusOK, pairDTO(userID, targetID, state))
	})
}

func writeNotImplemented(w http.ResponseWriter) {
	writeError(w, http.StatusNotImplemented, "not_implemented", "sprint 3 handler is not implemented")
}
