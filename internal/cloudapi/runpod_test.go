package cloudapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunPodSubmitWrapsInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", request.Method)
		}
		if request.URL.Path != "/v2/endpoint-123/run" {
			t.Fatalf("path = %s, want /v2/endpoint-123/run", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("Authorization = %q, want Bearer secret", got)
		}

		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		input, ok := payload["input"].(map[string]any)
		if !ok {
			t.Fatalf("payload = %#v, want input wrapper", payload)
		}
		if input["prompt"] != "hello" {
			t.Fatalf("prompt = %#v, want hello", input["prompt"])
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":"job-1","status":"IN_QUEUE"}`))
	}))
	defer server.Close()

	client, err := NewRunPodClient(server.Client(), "secret", "endpoint-123", server.URL+"/v2")
	if err != nil {
		t.Fatalf("NewRunPodClient() error = %v", err)
	}

	result, err := client.Submit(map[string]any{
		"input": map[string]any{"prompt": "hello"},
	}, false)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if result.RequestID != "job-1" {
		t.Fatalf("RequestID = %q, want job-1", result.RequestID)
	}
	if result.Status != "IN_QUEUE" {
		t.Fatalf("Status = %q, want IN_QUEUE", result.Status)
	}
}

func TestRunPodStatusUsesStatusEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", request.Method)
		}
		if request.URL.Path != "/v2/endpoint-123/status/job-1" {
			t.Fatalf("path = %s, want /v2/endpoint-123/status/job-1", request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":"job-1","status":"COMPLETED","output":{"ok":true}}`))
	}))
	defer server.Close()

	client, err := NewRunPodClient(server.Client(), "secret", "endpoint-123", server.URL+"/v2")
	if err != nil {
		t.Fatalf("NewRunPodClient() error = %v", err)
	}

	result, err := client.Result("job-1")
	if err != nil {
		t.Fatalf("Result() error = %v", err)
	}
	if result.Operation != "result" {
		t.Fatalf("Operation = %q, want result", result.Operation)
	}
	if result.Status != "COMPLETED" {
		t.Fatalf("Status = %q, want COMPLETED", result.Status)
	}
}
