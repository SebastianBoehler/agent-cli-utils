package youtubecontrol

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNormalizeOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default relevance omitted", input: "", want: ""},
		{name: "date preserved", input: "date", want: "date"},
		{name: "view count canonicalized", input: "viewCount", want: "viewCount"},
		{name: "video count canonicalized", input: "videoCount", want: "videoCount"},
		{name: "invalid rejected", input: "popular", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeOrder(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeOrder(%q) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeOrder(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeOrder(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSearchCanonicalizesOrderInResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.URL.Query().Get("order"); got != "viewCount" {
			t.Fatalf("request order = %q, want %q", got, "viewCount")
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	previousEndpoint := youtubeSearchEndpoint
	youtubeSearchEndpoint = server.URL
	defer func() {
		youtubeSearchEndpoint = previousEndpoint
	}()

	previousAPIKey, hadAPIKey := os.LookupEnv("YOUTUBE_API_KEY")
	if err := os.Setenv("YOUTUBE_API_KEY", "test-key"); err != nil {
		t.Fatalf("set YOUTUBE_API_KEY: %v", err)
	}
	defer func() {
		if hadAPIKey {
			_ = os.Setenv("YOUTUBE_API_KEY", previousAPIKey)
			return
		}
		_ = os.Unsetenv("YOUTUBE_API_KEY")
	}()

	service := NewService(server.Client())
	result, err := service.Search(context.Background(), SearchOptions{
		Query:   "lofi",
		Order:   "VIEWCOUNT",
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if result.Order != "viewCount" {
		t.Fatalf("Search().Order = %q, want %q", result.Order, "viewCount")
	}
}
