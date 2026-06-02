package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redsylx/tldr-discord-service/internal/model"
)

func TestSendTextMessage_Success(t *testing.T) {
	var got map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, 0)
	err := c.SendTextMessage(context.Background(), "hello world", "")
	if err != nil {
		t.Fatal(err)
	}
	if got["content"] != "hello world" {
		t.Errorf("expected content 'hello world', got %q", got["content"])
	}
}

func TestSendTextMessage_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, 0)
	err := c.SendTextMessage(context.Background(), "hello", "")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestSendEmbed_Success(t *testing.T) {
	var got map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, 0)
	embed := model.Embed{
		Title: "Test",
		Color: 0xFF0000,
		Fields: []model.Field{
			{Name: "Key", Value: "Val", Inline: true},
		},
	}
	err := c.SendEmbed(context.Background(), embed, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateThread_Success(t *testing.T) {
	var got map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("wait") != "true" {
			t.Error("expected wait=true query param")
		}
		json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"channel_id": "1511342083150053463",
		})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, 0)
	threadID, err := c.CreateThread(context.Background(), "Test Title", "Test Desc", "https://example.com", 5, []string{"1511311266935738478"})
	if err != nil {
		t.Fatal(err)
	}

	if threadID != "1511342083150053463" {
		t.Errorf("expected 1511342083150053463, got %q", threadID)
	}

	if got["thread_name"] != "Test Title" {
		t.Errorf("expected 'Test Title', got %q", got["thread_name"])
	}

	embeds, ok := got["embeds"].([]any)
	if !ok || len(embeds) != 1 {
		t.Fatal("expected 1 embed")
	}
	embed := embeds[0].(map[string]any)
	if embed["title"] != "Test Title" {
		t.Errorf("expected 'Test Title', got %q", embed["title"])
	}
	if embed["url"] != "https://example.com" {
		t.Errorf("expected 'https://example.com', got %q", embed["url"])
	}
	if embed["description"] != "Test Desc" {
		t.Errorf("expected 'Test Desc', got %q", embed["description"])
	}

	appliedTags, ok := got["applied_tags"].([]any)
	if !ok || len(appliedTags) != 1 {
		t.Fatal("expected 1 applied tag")
	}
	if appliedTags[0] != "1511311266935738478" {
		t.Errorf("expected tag ID, got %v", appliedTags[0])
	}
}

func TestCreateThread_NoTags(t *testing.T) {
	var got map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"channel_id": "1511342083150053463",
		})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, 0)
	_, err := c.CreateThread(context.Background(), "No Tags", "Desc", "", 0, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := got["applied_tags"]; ok {
		t.Error("expected no applied_tags")
	}

	embeds, ok := got["embeds"].([]any)
	if !ok || len(embeds) != 1 {
		t.Fatal("expected 1 embed")
	}
	embed := embeds[0].(map[string]any)
	if embed["title"] != "No Tags" {
		t.Errorf("expected 'No Tags', got %q", embed["title"])
	}
	if _, ok := embed["fields"]; ok {
		t.Error("expected no fields when read=0")
	}
}

func TestCreateThread_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, 0)
	_, err := c.CreateThread(context.Background(), "title", "desc", "", 0, nil)
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}