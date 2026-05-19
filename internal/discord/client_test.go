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
	err := c.SendTextMessage(context.Background(), "hello world")
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
	err := c.SendTextMessage(context.Background(), "hello")
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
	err := c.SendEmbed(context.Background(), embed)
	if err != nil {
		t.Fatal(err)
	}
}