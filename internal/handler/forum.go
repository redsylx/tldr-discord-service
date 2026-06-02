package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/redsylx/tldr-discord-service/internal/model"
)

func (h *Handler) HandleForum(w http.ResponseWriter, r *http.Request) {
	var req model.CreateForumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := model.ValidateCreateForumRequest(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	threadID, err := h.client.CreateThread(r.Context(), req.Title, req.Description, req.URL, req.Read, req.Tags)
	if err != nil {
		slog.Error("forum handler error", "err", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ForumResponse{ThreadID: threadID})
}
