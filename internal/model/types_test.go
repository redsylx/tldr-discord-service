package model

import "testing"

func TestValidateFailedPayload_Valid(t *testing.T) {
	p := FailedPayload{
		ProcessName: "test-process",
		TraceInfo: TraceInfo{
			EmailId:   "test@example.com",
			NewsTitle: "Test News",
			NewsUrl:   "https://example.com",
		},
		Message: "something went wrong",
	}
	if err := ValidateFailedPayload(p); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateFailedPayload_ValidWithoutNewsUrl(t *testing.T) {
	p := FailedPayload{
		ProcessName: "test-process",
		TraceInfo: TraceInfo{
			EmailId:   "test@example.com",
			NewsTitle: "Test News",
		},
		Message: "something went wrong",
	}
	if err := ValidateFailedPayload(p); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateFailedPayload_MissingProcessName(t *testing.T) {
	p := FailedPayload{
		TraceInfo: TraceInfo{EmailId: "e", NewsTitle: "t"},
		Message:   "m",
	}
	if err := ValidateFailedPayload(p); err == nil {
		t.Error("expected error for missing processName")
	}
}

func TestValidateFailedPayload_MissingEmailId(t *testing.T) {
	p := FailedPayload{
		ProcessName: "p",
		TraceInfo:   TraceInfo{NewsTitle: "t"},
		Message:     "m",
	}
	if err := ValidateFailedPayload(p); err == nil {
		t.Error("expected error for missing emailId")
	}
}

func TestValidateFailedPayload_MissingNewsTitle(t *testing.T) {
	p := FailedPayload{
		ProcessName: "p",
		TraceInfo:   TraceInfo{EmailId: "e"},
		Message:     "m",
	}
	if err := ValidateFailedPayload(p); err == nil {
		t.Error("expected error for missing newsTitle")
	}
}

func TestValidateFailedPayload_MissingMessage(t *testing.T) {
	p := FailedPayload{
		ProcessName: "p",
		TraceInfo:   TraceInfo{EmailId: "e", NewsTitle: "t"},
	}
	if err := ValidateFailedPayload(p); err == nil {
		t.Error("expected error for missing message")
	}
}