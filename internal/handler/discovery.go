package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type ContextKey string

const CtxKeyUserID ContextKey = "current_user_id" //temp, will need to see what is used in middleware when it's availible

func DiscoveryHandler(disc Discovery) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "invalid_request", "GET required")
			return
		}

		userID, ok := r.Context().Value(CtxKeyUserID).(int)
		if !ok || userID <= 0 {
			writeError(w, http.StatusBadRequest, "invalid_request", "no known user id")
			return
		}

		query := r.URL.Query()
		allowedParams := map[string]bool{"interest": true, "limit": true, "offset": true}

		for key, vals := range query {
			if !allowedParams[key] {
				writeError(w, http.StatusBadRequest, "invalid_request", "unknown query parameter: "+key)
				return
			}
			if len(vals) > 1 {
				writeError(w, http.StatusBadRequest, "invalid_request", "repeated query parameter: "+key)
				return
			}
		}

		var interestFilter string
		if val := query.Get("interest"); val != "" {
			interestFilter = strings.TrimSpace(val)
		}

		limit := 20
		if val := query.Get("limit"); val != "" {
			o_Limit, err := strconv.Atoi(val)
			if err != nil || o_Limit < 1 || o_Limit > 50 {
				writeError(w, http.StatusBadRequest, "invalid_request", "invalid limit")
				return
			}
			limit = o_Limit
		}

		offset := 0
		if val := query.Get("offset"); val != "" {
			o_Offset, err := strconv.Atoi(val)
			if err != nil || o_Offset < 0 {
				writeError(w, http.StatusBadRequest, "invalid_request", "invalid offset")
				return
			}
			offset = o_Offset
		}

		profiles, err := disc.Discover(r.Context(), userID)
		if err != nil {
			handleDiscoveryError(w, err)
			return
		}

		filtered := make([]domain.Profile, 0, len(profiles))
		for _, p := range profiles {
			if interestFilter == "" {
				filtered = append(filtered, p)
				continue
			}

			matched := false
			for _, inter := range p.Interests {
				if strings.EqualFold(strings.TrimSpace(string(inter)), interestFilter) {
					matched = true
					break
				}
			}
			if matched {
				filtered = append(filtered, p)
			}
		}

		var paged []domain.Profile
		if offset >= len(filtered) {
			paged = []domain.Profile{}
		} else {
			end := offset + limit
			if end > len(filtered) {
				end = len(filtered)
			}
			paged = filtered[offset:end]
		}

		Items := make([]DiscoveryItem, 0, len(paged))
		for _, p := range paged {
			ints := make([]string, len(p.Interests))
			for i, inter := range p.Interests {
				ints[i] = string(inter)
			}
			Items = append(Items, DiscoveryItem{
				UserID:    p.UserID,
				City:      p.City,
				Interests: ints,
			})
		}

		resp := DiscoveryResponse{
			Items:  Items,
			Limit:  limit,
			Offset: offset,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	})
}

func handleDiscoveryError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrUserNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
}
