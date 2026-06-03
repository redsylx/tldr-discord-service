package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/redsylx/tldr-discord-service/internal/model"
)

func (h *Handler) HandleProcess(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if text == "" {
		respondError(w, http.StatusBadRequest, "text is required")
		return
	}

	jsonPath := text + ".json"
	var req model.CreateForumRequest
	if err := h.reader.ReadJSON(r.Context(), jsonPath, &req); err != nil {
		slog.Error("process handler: read json", "path", jsonPath, "err", err)
		respondError(w, http.StatusInternalServerError, "failed to read forum data: "+err.Error())
		return
	}

	if err := model.ValidateCreateForumRequest(req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid forum data: "+err.Error())
		return
	}

	threadID, err := h.client.CreateThread(r.Context(), req.Title, req.Description, req.URL, req.Read, req.Tags)
	if err != nil {
		slog.Error("process handler: create thread", "err", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	txtPath := text + ".txt"
	if err := h.processTextFile(r.Context(), txtPath, threadID); err != nil {
		slog.Error("process handler: process text", "path", txtPath, "err", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) processTextFile(ctx context.Context, path string, threadID string) error {
	batchSize := h.cfg.BatchLineCount
	if batchSize <= 0 {
		batchSize = 5
	}

	var batch []string
	hasLines := false

	err := h.reader.StreamLines(ctx, path, func(line string) error {
		hasLines = true
		batch = append(batch, line)
		if len(batch) >= batchSize {
			content := strings.Join(batch, "\n")
			batch = batch[:0]
			return h.client.SendTextMessage(ctx, content, threadID)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if !hasLines {
		return h.client.SendTextMessage(ctx, "File is empty", threadID)
	}

	if len(batch) > 0 {
		content := strings.Join(batch, "\n")
		return h.client.SendTextMessage(ctx, content, threadID)
	}

	return nil
}
