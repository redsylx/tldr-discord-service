package model

import "errors"

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