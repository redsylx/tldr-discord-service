package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/redsylx/tldr-discord-service/internal/discord"
	"github.com/redsylx/tldr-discord-service/internal/gcs"
	"github.com/redsylx/tldr-discord-service/internal/model"
)

type Handler struct {
	reader  gcs.Reader
	client  discord.Client
	cfg     model.Config
}

func New(reader gcs.Reader, client discord.Client, cfg model.Config) *Handler {
	return &Handler{reader: reader, client: client, cfg: cfg}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	file := r.URL.Query().Get("file")
	typ := r.URL.Query().Get("type")

	if file == "" {
		respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	if typ != "success" && typ != "failed" {
		respondError(w, http.StatusBadRequest, "type must be success or failed")
		return
	}

	var err error
	if typ == "success" {
		err = h.handleSuccess(r.Context(), file)
	} else {
		err = h.handleFailed(r.Context(), file)
	}

	if err != nil {
		slog.Error("handler error", "file", file, "type", typ, "err", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleSuccess(ctx context.Context, file string) error {
	lines, err := h.reader.ReadTextLines(ctx, file)
	if err != nil {
		return err
	}

	if len(lines) == 0 {
		return h.client.SendTextMessage(ctx, "File is empty")
	}

	batch := h.cfg.BatchLineCount
	if batch <= 0 {
		batch = 5
	}

	for i := 0; i < len(lines); i += batch {
		end := i + batch
		if end > len(lines) {
			end = len(lines)
		}

		var content string
		for j, line := range lines[i:end] {
			if j > 0 {
				content += "\n"
			}
			content += line
		}

		if err := h.client.SendTextMessage(ctx, content); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) handleFailed(ctx context.Context, file string) error {
	var payload model.FailedPayload
	if err := h.reader.ReadJSON(ctx, file, &payload); err != nil {
		return err
	}

	if err := model.ValidateFailedPayload(payload); err != nil {
		return err
	}

	embed := model.Embed{
		Title: "❌ Process Failed: " + payload.ProcessName,
		Color: 0xE74C3C,
		Fields: []model.Field{
			{Name: "Email ID", Value: payload.TraceInfo.EmailId, Inline: true},
			{Name: "News Title", Value: payload.TraceInfo.NewsTitle, Inline: true},
		},
	}

	if payload.TraceInfo.NewsUrl != "" {
		embed.Fields = append(embed.Fields, model.Field{
			Name: "News URL", Value: payload.TraceInfo.NewsUrl, Inline: false,
		})
	}

	embed.Fields = append(embed.Fields, model.Field{
		Name: "Error Message", Value: payload.Message, Inline: false,
	})

	return h.client.SendEmbed(ctx, embed)
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}