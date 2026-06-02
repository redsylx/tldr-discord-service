package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/redsylx/tldr-discord-service/internal/model"
)

type mockGCSReader struct {
	lines  []string
	jsonData any
	readErr error
}

func (m *mockGCSReader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return nil, m.readErr
}

func (m *mockGCSReader) ReadTextLines(ctx context.Context, path string) ([]string, error) {
	return m.lines, m.readErr
}

func (m *mockGCSReader) ReadJSON(ctx context.Context, path string, dst any) error {
	if m.readErr != nil {
		return m.readErr
	}
	data, _ := json.Marshal(m.jsonData)
	return json.Unmarshal(data, dst)
}

type mockDiscordClient struct {
	sentTexts  []string
	sentEmbeds []model.Embed
	createdThreads []createThreadCall
	sendErr    error
}

type createThreadCall struct {
	Title       string
	Description string
	Tags        []string
}

func (m *mockDiscordClient) SendTextMessage(ctx context.Context, content string, threadID string) error {
	m.sentTexts = append(m.sentTexts, content)
	return m.sendErr
}

func (m *mockDiscordClient) SendEmbed(ctx context.Context, embed model.Embed, threadID string) error {
	m.sentEmbeds = append(m.sentEmbeds, embed)
	return m.sendErr
}

func (m *mockDiscordClient) CreateThread(ctx context.Context, title string, description string, tags []string) (string, error) {
	m.createdThreads = append(m.createdThreads, createThreadCall{
		Title: title, Description: description, Tags: tags,
	})
	return "thread-123", m.sendErr
}

func cfg() model.Config {
	return model.Config{
		BatchLineCount: 2,
		DiscordDelayMs: 0,
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := New(&mockGCSReader{}, &mockDiscordClient{}, cfg())
	req := httptest.NewRequest(http.MethodGet, "/webhook?file=test.txt&type=success", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestHandler_MissingFile(t *testing.T) {
	h := New(&mockGCSReader{}, &mockDiscordClient{}, cfg())
	req := httptest.NewRequest(http.MethodPost, "/webhook?type=success&thread_id=123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_InvalidType(t *testing.T) {
	h := New(&mockGCSReader{}, &mockDiscordClient{}, cfg())
	req := httptest.NewRequest(http.MethodPost, "/webhook?file=test.txt&type=invalid&thread_id=123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_Success_SendsBatches(t *testing.T) {
	mock := &mockDiscordClient{}
	h := New(&mockGCSReader{lines: []string{"a", "b", "c", "d", "e"}}, mock, cfg())

	req := httptest.NewRequest(http.MethodPost, "/webhook?file=test.txt&type=success&thread_id=123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
	}

	if len(mock.sentTexts) != 3 {
		t.Fatalf("expected 3 batches, got %d: %v", len(mock.sentTexts), mock.sentTexts)
	}
	expected := []string{"a\nb", "c\nd", "e"}
	for i, e := range expected {
		if mock.sentTexts[i] != e {
			t.Errorf("batch %d: expected %q, got %q", i, e, mock.sentTexts[i])
		}
	}
}

func TestHandler_Success_EmptyFile(t *testing.T) {
	mock := &mockDiscordClient{}
	h := New(&mockGCSReader{lines: []string{}}, mock, cfg())

	req := httptest.NewRequest(http.MethodPost, "/webhook?file=empty.txt&type=success&thread_id=123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
	}

	if len(mock.sentTexts) != 1 {
		t.Fatalf("expected 1 message, got %d", len(mock.sentTexts))
	}
	if mock.sentTexts[0] != "File is empty" {
		t.Errorf("expected 'File is empty', got %q", mock.sentTexts[0])
	}
}

func TestHandler_Success_GCSError(t *testing.T) {
	h := New(&mockGCSReader{readErr: errors.New("gcs error")}, &mockDiscordClient{}, cfg())

	req := httptest.NewRequest(http.MethodPost, "/webhook?file=test.txt&type=success&thread_id=123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandler_Failed_SendsEmbed(t *testing.T) {
	payload := model.FailedPayload{
		ProcessName: "test-process",
		TraceInfo: model.TraceInfo{
			EmailId:   "e@e.com",
			NewsTitle: "Title",
			NewsUrl:   "https://url",
		},
		Message: "error msg",
	}

	mock := &mockDiscordClient{}
	h := New(&mockGCSReader{jsonData: payload}, mock, cfg())

	req := httptest.NewRequest(http.MethodPost, "/webhook?file=fail.json&type=failed&thread_id=123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	if len(mock.sentEmbeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(mock.sentEmbeds))
	}

	embed := mock.sentEmbeds[0]
	if !strings.Contains(embed.Title, "test-process") {
		t.Errorf("expected title to contain process name, got %q", embed.Title)
	}
	if embed.Color != 0xE74C3C {
		t.Errorf("expected color 0xE74C3C, got %d", embed.Color)
	}
	if len(embed.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(embed.Fields))
	}
}

func TestHandler_Failed_ValidationError(t *testing.T) {
	mock := &mockDiscordClient{}
	h := New(&mockGCSReader{jsonData: model.FailedPayload{}}, mock, cfg())

	req := httptest.NewRequest(http.MethodPost, "/webhook?file=fail.json&type=failed&thread_id=123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if len(mock.sentEmbeds) != 0 {
		t.Error("expected no embed sent for invalid payload")
	}
}

func TestHandler_Failed_GCSError(t *testing.T) {
	mock := &mockDiscordClient{}
	h := New(&mockGCSReader{readErr: errors.New("gcs error")}, mock, cfg())

	req := httptest.NewRequest(http.MethodPost, "/webhook?file=fail.json&type=failed&thread_id=123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if len(mock.sentEmbeds) != 0 {
		t.Error("expected no embed sent on GCS error")
	}
}

func TestForumHandler_Success(t *testing.T) {
	mock := &mockDiscordClient{}
	h := New(&mockGCSReader{}, mock, cfg())

	body := `{"title":"Test Post","desc":"Test description","tags":["tag1","tag2"]}`
	req := httptest.NewRequest(http.MethodPost, "/forum", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.HandleForum(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp model.ForumResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.ThreadID != "thread-123" {
		t.Errorf("expected thread-123, got %q", resp.ThreadID)
	}

	if len(mock.createdThreads) != 1 {
		t.Fatalf("expected 1 thread creation, got %d", len(mock.createdThreads))
	}
	if mock.createdThreads[0].Title != "Test Post" {
		t.Errorf("expected 'Test Post', got %q", mock.createdThreads[0].Title)
	}
	if len(mock.createdThreads[0].Tags) != 2 || mock.createdThreads[0].Tags[0] != "tag1" {
		t.Errorf("unexpected tags: %v", mock.createdThreads[0].Tags)
	}
}

func TestForumHandler_MissingTitle(t *testing.T) {
	h := New(&mockGCSReader{}, &mockDiscordClient{}, cfg())

	body := `{"desc":"Test description","tags":["tag1"]}`
	req := httptest.NewRequest(http.MethodPost, "/forum", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.HandleForum(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestForumHandler_MissingDescription(t *testing.T) {
	h := New(&mockGCSReader{}, &mockDiscordClient{}, cfg())

	body := `{"title":"Test Post","tags":["tag1"]}`
	req := httptest.NewRequest(http.MethodPost, "/forum", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.HandleForum(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestForumHandler_InvalidJSON(t *testing.T) {
	h := New(&mockGCSReader{}, &mockDiscordClient{}, cfg())

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/forum", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.HandleForum(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestForumHandler_ClientError(t *testing.T) {
	mock := &mockDiscordClient{sendErr: errors.New("discord error")}
	h := New(&mockGCSReader{}, mock, cfg())

	body := `{"title":"Test Post","desc":"Test description","tags":["tag1"]}`
	req := httptest.NewRequest(http.MethodPost, "/forum", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.HandleForum(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}