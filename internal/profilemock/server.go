package profilemock

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	mu   sync.RWMutex
	mode string
}

func NewHandler() http.Handler {
	server := &Server{
		mode: "ok",
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /v1/profiles/{userID}",
		server.getProfile,
	)

	mux.HandleFunc(
		"PUT /__admin/mode/{mode}",
		server.setMode,
	)

	return mux
}

func (s *Server) getMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.mode
}

func (s *Server) setMode(w http.ResponseWriter, r *http.Request) {
	mode := r.PathValue("mode")

	switch mode {
	case "ok", "missing", "error", "slow", "large":
	default:
		writeError(w, http.StatusBadRequest, "invalid mode")
		return
	}

	s.mu.Lock()
	s.mode = mode
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{
		"mode": mode,
	})
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("userID")

	if rawID == "" || rawID[0] == '0' {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	for _, ch := range rawID {
		if ch < '0' || ch > '9' {
			writeError(w, http.StatusBadRequest, "invalid user ID")
			return
		}
	}

	userID, err := strconv.Atoi(rawID)
	if err != nil || userID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	mode := s.getMode()

	switch mode {
	case "missing":
		writeError(w, http.StatusNotFound, "profile not found")
		return

	case "error":
		writeError(w, http.StatusInternalServerError, "internal error")
		return

	case "slow":
		timer := time.NewTimer(2 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-r.Context().Done():
			return
		}
	}

	profile, found := demoProfile(userID)
	if !found {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}

	if mode == "large" {
		profile.City = strings.Repeat("M", 128<<10)
	}

	writeJSON(w, http.StatusOK, profile)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
