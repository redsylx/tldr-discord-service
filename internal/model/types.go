package model

import (
	"errors"
	"fmt"
)

type Config struct {
	Port            string
	GCSBucket       string
	DiscordWebhook  string
	BatchLineCount  int
	DiscordDelayMs  int
}

type WebhookParams struct {
	File     string
	Type     string
	ThreadID string
}

type TraceInfo struct {
	EmailId   string `json:"emailId"`
	NewsTitle string `json:"newsTitle"`
	NewsUrl   string `json:"newsUrl,omitempty"`
}

type FailedPayload struct {
	ProcessName string    `json:"processName"`
	TraceInfo   TraceInfo `json:"traceInfo"`
	Message     string    `json:"message"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type Embed struct {
	Title  string  `json:"title"`
	Color  int     `json:"color"`
	Fields []Field `json:"fields"`
}

type CreateForumRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"desc"`
	Tags        []string `json:"tags"`
}

type ForumResponse struct {
	ThreadID string `json:"thread_id"`
}

var TagPresets = map[string]string{
	"p1": "1511311266935738478",
	"p2": "1511311288922275840",
	"p3": "1511314922540240936",
	"p4": "1511314957919195268",
}

var ValidTagIDs map[string]bool

func init() {
	ValidTagIDs = make(map[string]bool, len(TagPresets))
	for _, id := range TagPresets {
		ValidTagIDs[id] = true
	}
}

func ValidateCreateForumRequest(req CreateForumRequest) error {
	if req.Title == "" {
		return errors.New("title is required")
	}
	if req.Description == "" {
		return errors.New("desc is required")
	}
	if len(req.Tags) > 4 {
		return errors.New("max 4 tags allowed")
	}
	for _, tag := range req.Tags {
		if !ValidTagIDs[tag] {
			return fmt.Errorf("invalid tag ID: %s", tag)
		}
	}
	return nil
}

func ValidateFailedPayload(p FailedPayload) error {
	if p.ProcessName == "" {
		return errors.New("processName is required")
	}
	if p.TraceInfo.EmailId == "" {
		return errors.New("traceInfo.emailId is required")
	}
	if p.TraceInfo.NewsTitle == "" {
		return errors.New("traceInfo.newsTitle is required")
	}
	if p.Message == "" {
		return errors.New("message is required")
	}
	return nil
}