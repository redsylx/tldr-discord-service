package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cloud.google.com/go/storage"

	"github.com/redsylx/tldr-discord-service/internal/discord"
	"github.com/redsylx/tldr-discord-service/internal/gcs"
	"github.com/redsylx/tldr-discord-service/internal/handler"
	"github.com/redsylx/tldr-discord-service/internal/model"
)

func loadConfig() model.Config {
	return model.Config{
		Port:           getEnv("PORT", "8080"),
		GCSBucket:      os.Getenv("GCS_BUCKET_NAME"),
		DiscordWebhook: os.Getenv("DISCORD_WEBHOOK_URL"),
		BatchLineCount: getEnvInt("BATCH_LINE_COUNT", 5),
		DiscordDelayMs: getEnvInt("DISCORD_DELAY_MS", 2000),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &loggingWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(lw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", lw.code,
			"duration", time.Since(start).String(),
		)
	})
}

type loggingWriter struct {
	http.ResponseWriter
	code int
}

func (w *loggingWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "panic", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := loadConfig()

	if cfg.GCSBucket == "" {
		slog.Error("GCS_BUCKET_NAME is required")
		os.Exit(1)
	}
	if cfg.DiscordWebhook == "" {
		slog.Error("DISCORD_WEBHOOK_URL is required")
		os.Exit(1)
	}

	storageClient, err := storage.NewClient(context.Background())
	if err != nil {
		slog.Error("failed to create storage client", "err", err)
		os.Exit(1)
	}
	defer storageClient.Close()

	gcsReader := gcs.NewReader(storageClient, cfg.GCSBucket)
	discordClient := discord.NewClient(cfg.DiscordWebhook, cfg.DiscordDelayMs)
	webhookHandler := handler.New(gcsReader, discordClient, cfg)

	mux := http.NewServeMux()
	mux.Handle("/webhook", webhookHandler)
	mux.HandleFunc("POST /forum", webhookHandler.HandleForum)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: recoveryMiddleware(loggingMiddleware(mux)),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	slog.Info("server stopped")
}