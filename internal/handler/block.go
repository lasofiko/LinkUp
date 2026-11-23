package handler

import "net/http"

// BlockHandler blocks a pair using the same user identity as the like handler.
func BlockHandler(blocks Blocks) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, targetID, ok := requestIDs(r)
		if !ok || !emptyBody(r) {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid request")
			return
		}
		if actorID == targetID {
			writeError(w, http.StatusConflict, "invalid_transition", "cannot block yourself")
			return
		}
		state, err := blocks.Block(r.Context(), actorID, targetID)
		if err != nil {
			writeHandlerError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, BlockResponse{TargetID: targetID, Blocked: state.Blocked, Mutual: state.IsMutual()})
	})
}
