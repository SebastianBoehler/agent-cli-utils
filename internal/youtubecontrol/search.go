package youtubecontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const youtubeSearchEndpoint = "https://www.googleapis.com/youtube/v3/search"

type youtubeSearchResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title        string `json:"title"`
			ChannelTitle string `json:"channelTitle"`
			PublishedAt  string `json:"publishedAt"`
			Description  string `json:"description"`
		} `json:"snippet"`
	} `json:"items"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (service *Service) Search(ctx context.Context, options SearchOptions) (SearchResult, error) {
	key := strings.TrimSpace(os.Getenv("YOUTUBE_API_KEY"))
	if key == "" {
		return SearchResult{}, fmt.Errorf("YOUTUBE_API_KEY is required")
	}

	query := strings.TrimSpace(options.Query)
	if query == "" {
		return SearchResult{}, fmt.Errorf("query is required")
	}

	ctx, cancel := withTimeout(ctx, options.Timeout, 10*time.Second)
	defer cancel()

	values := url.Values{}
	values.Set("part", "snippet")
	values.Set("type", "video")
	values.Set("q", query)
	values.Set("key", key)
	values.Set("maxResults", fmt.Sprintf("%d", clampMaxResults(options.MaxResults)))
	if value := normalizeLanguage(options.Language); value != "" {
		values.Set("relevanceLanguage", value)
	}
	if value := normalizeRegion(options.Region); value != "" {
		values.Set("regionCode", value)
	}
	if value, err := normalizeDuration(options.Duration); err != nil {
		return SearchResult{}, err
	} else if value != "" {
		values.Set("videoDuration", value)
	}
	if value, err := normalizeCaption(options.Caption); err != nil {
		return SearchResult{}, err
	} else if value != "" {
		values.Set("videoCaption", value)
	}
	if value, err := normalizeOrder(options.Order); err != nil {
		return SearchResult{}, err
	} else if value != "" {
		values.Set("order", value)
	}
	if value, err := normalizeSafeSearch(options.SafeSearch); err != nil {
		return SearchResult{}, err
	} else if value != "" {
		values.Set("safeSearch", value)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, youtubeSearchEndpoint+"?"+values.Encode(), nil)
	if err != nil {
		return SearchResult{}, err
	}
	response, err := service.client.Do(request)
	if err != nil {
		return SearchResult{}, err
	}
	defer response.Body.Close()

	var payload youtubeSearchResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return SearchResult{}, fmt.Errorf("decode youtube search response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return SearchResult{}, fmt.Errorf("youtube search failed: %s", firstNonEmpty(payloadError(payload), response.Status))
	}

	result := SearchResult{
		Query:      query,
		Language:   normalizeLanguage(options.Language),
		Region:     normalizeRegion(options.Region),
		Duration:   strings.TrimSpace(options.Duration),
		Caption:    strings.TrimSpace(options.Caption),
		Order:      strings.TrimSpace(options.Order),
		SafeSearch: strings.TrimSpace(options.SafeSearch),
		MaxResults: clampMaxResults(options.MaxResults),
		Items:      make([]SearchItem, 0, len(payload.Items)),
	}
	for _, item := range payload.Items {
		if strings.TrimSpace(item.ID.VideoID) == "" {
			continue
		}
		result.Items = append(result.Items, SearchItem{
			VideoID:      item.ID.VideoID,
			Title:        item.Snippet.Title,
			ChannelTitle: item.Snippet.ChannelTitle,
			PublishedAt:  item.Snippet.PublishedAt,
			Description:  item.Snippet.Description,
			URL:          "https://www.youtube.com/watch?v=" + item.ID.VideoID,
		})
	}
	return result, nil
}

func withTimeout(ctx context.Context, timeout time.Duration, fallback time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = fallback
	}
	return context.WithTimeout(ctx, timeout)
}

func clampMaxResults(value int) int {
	if value <= 0 {
		return 10
	}
	if value > 50 {
		return 50
	}
	return value
}

func normalizeLanguage(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeRegion(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeDuration(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "any":
		return "", nil
	case "short", "medium", "long":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported duration %q", value)
	}
}

func normalizeCaption(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "any":
		return "", nil
	case "closedcaption", "closed_caption", "closed-caption":
		return "closedCaption", nil
	case "none":
		return "none", nil
	default:
		return "", fmt.Errorf("unsupported caption filter %q", value)
	}
}

func normalizeOrder(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "relevance":
		return "", nil
	case "date", "rating", "title":
		return value, nil
	case "videocount":
		return "videoCount", nil
	case "viewcount":
		return "viewCount", nil
	default:
		return "", fmt.Errorf("unsupported order %q", value)
	}
}

func normalizeSafeSearch(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "moderate":
		return "", nil
	case "none", "strict":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported safe search %q", value)
	}
}

func payloadError(payload youtubeSearchResponse) string {
	if payload.Error == nil {
		return ""
	}
	return strings.TrimSpace(payload.Error.Message)
}
