package youtubecontrol

import "testing"

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
