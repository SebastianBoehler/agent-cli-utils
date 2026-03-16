package tvcontrol

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlayAirPlay(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/play" {
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		payload, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		capturedBody = string(payload)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result, err := playAirPlay(context.Background(), server.Client(), Device{
		Name:     "Living Room",
		Protocol: ProtocolAirPlay,
		Location: server.URL,
	}, "http://example.test/video.m3u8", 0.25)
	if err != nil {
		t.Fatalf("play airplay: %v", err)
	}

	if !result.OK {
		t.Fatalf("expected OK result: %+v", result)
	}
	if !strings.Contains(capturedBody, "Content-Location: http://example.test/video.m3u8") {
		t.Fatalf("expected content location in body, got %q", capturedBody)
	}
	if !strings.Contains(capturedBody, "Start-Position: 0.250") {
		t.Fatalf("expected start position in body, got %q", capturedBody)
	}
}
