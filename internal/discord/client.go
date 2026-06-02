package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/redsylx/tldr-discord-service/internal/model"
)

type Client interface {
	SendTextMessage(ctx context.Context, content string, threadID string) error
	SendEmbed(ctx context.Context, embed model.Embed, threadID string) error
}

func NewClient(webhookURL string, delayMs int) Client {
	return &discordClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		delay:      time.Duration(delayMs) * time.Millisecond,
	}
}

type discordClient struct {
	webhookURL string
	httpClient *http.Client
	delay      time.Duration
}

func (c *discordClient) SendTextMessage(ctx context.Context, content string, threadID string) error {
	payload := map[string]string{"content": content}
	return c.send(ctx, payload, threadID)
}

func (c *discordClient) SendEmbed(ctx context.Context, embed model.Embed, threadID string) error {
	payload := map[string]any{"embeds": []model.Embed{embed}}
	return c.send(ctx, payload, threadID)
}

func (c *discordClient) send(ctx context.Context, payload any, threadID string) error {
	defer func() {
		if c.delay > 0 {
			time.Sleep(c.delay)
		}
	}()

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: marshal payload: %w", err)
	}

	urlStr := c.webhookURL
	if threadID != "" {
		u, parseErr := url.Parse(c.webhookURL)
		if parseErr == nil {
			q := u.Query()
			q.Set("thread_id", threadID)
			u.RawQuery = q.Encode()
			urlStr = u.String()
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}

	return nil
}