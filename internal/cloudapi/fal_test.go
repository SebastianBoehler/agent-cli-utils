package cloudapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFalSubmitUsesQueueEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", request.Method)
		}
		if request.URL.Path != "/fal-ai/flux/dev" {
			t.Fatalf("path = %s, want /fal-ai/flux/dev", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Key secret" {
			t.Fatalf("Authorization = %q, want Key secret", got)
		}

		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["prompt"] != "hello" {
			t.Fatalf("prompt = %#v, want hello", payload["prompt"])
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"request_id":"req-1","status":"IN_QUEUE"}`))
	}))
	defer server.Close()

	client, err := NewFalClient(server.Client(), "secret", "fal-ai/flux/dev", server.URL)
	if err != nil {
		t.Fatalf("NewFalClient() error = %v", err)
	}

	result, err := client.Submit(map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if result.RequestID != "req-1" {
		t.Fatalf("RequestID = %q, want req-1", result.RequestID)
	}
	if result.Status != "IN_QUEUE" {
		t.Fatalf("Status = %q, want IN_QUEUE", result.Status)
	}
}

func TestFalStatusAddsLogsQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", request.Method)
		}
		if request.URL.Path != "/fal-ai/flux/dev/requests/req-1/status" {
			t.Fatalf("path = %s, want /fal-ai/flux/dev/requests/req-1/status", request.URL.Path)
		}
		if request.URL.Query().Get("logs") != "1" {
			t.Fatalf("logs query = %q, want 1", request.URL.Query().Get("logs"))
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"status":"COMPLETED"}`))
	}))
	defer server.Close()

	client, err := NewFalClient(server.Client(), "secret", "fal-ai/flux/dev", server.URL)
	if err != nil {
		t.Fatalf("NewFalClient() error = %v", err)
	}

	result, err := client.Status("req-1", true)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if result.Status != "COMPLETED" {
		t.Fatalf("Status = %q, want COMPLETED", result.Status)
	}
}
