package company

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"
	"time"
)

type HandelsregisterProvider struct {
	client    *http.Client
	baseURL   string
	userAgent string
	delay     time.Duration
}

func NewHandelsregisterProvider(cfg Config) *HandelsregisterProvider {
	client, err := newCookieClient(cfg.HTTPClient, cfg.Timeout)
	if err != nil {
		client = cloneHTTPClient(cfg.HTTPClient, cfg.Timeout)
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.HandelsregisterURL), "/")
	if baseURL == "" {
		baseURL = defaultHandelsregister
	}

	userAgent := strings.TrimSpace(cfg.UserAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	delay := cfg.HandelsregisterDelay
	if delay <= 0 {
		delay = 250 * time.Millisecond
	}

	return &HandelsregisterProvider{
		client:    client,
		baseURL:   baseURL,
		userAgent: userAgent,
		delay:     delay,
	}
}

func (provider *HandelsregisterProvider) Name() string {
	return SourceHandelsregister
}

func (provider *HandelsregisterProvider) Search(ctx context.Context, req SearchRequest) ([]CompanyResult, error) {
	if req.Country != defaultCountry {
		return nil, fmt.Errorf("handelsregister only supports country=%s", defaultCountry)
	}

	pageURL := provider.baseURL + "/rp_web/erweitertesuche/welcome.xhtml"
	viewState, err := provider.fetchViewState(ctx, pageURL)
	if err != nil {
		return nil, err
	}

	if err := waitForRateLimit(ctx, provider.delay); err != nil {
		return nil, err
	}

	values := neturl.Values{
		"form":                          {"form"},
		"form:schlagwoerter":            {req.Query},
		"form:schlagwortOptionen":       {handelsregisterKeywordMode(req.Exact)},
		"form:ergebnisseProSeite_input": {handelsregisterPageSize(req.Limit)},
		"form:btnSuche":                 {"form:btnSuche"},
		"javax.faces.ViewState":         {viewState},
	}
	if req.City != "" {
		values.Set("form:ort", req.City)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, pageURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Referer", pageURL)
	request.Header.Set("User-Agent", provider.userAgent)

	response, err := provider.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	body, err := readResponse(response)
	if err != nil {
		return nil, err
	}

	results, err := parseHandelsregisterResults(body, pageURL, provider.baseURL)
	if err != nil {
		return nil, err
	}
	if len(results) > req.Limit {
		results = results[:req.Limit]
	}
	return results, nil
}

func (provider *HandelsregisterProvider) fetchViewState(ctx context.Context, pageURL string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("build bootstrap request: %w", err)
	}
	request.Header.Set("User-Agent", provider.userAgent)

	response, err := provider.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("bootstrap request failed: %w", err)
	}

	body, err := readResponse(response)
	if err != nil {
		return "", err
	}

	viewState, err := extractHandelsregisterViewState(body)
	if err != nil {
		return "", err
	}
	return viewState, nil
}

func waitForRateLimit(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func handelsregisterKeywordMode(exact bool) string {
	if exact {
		return "3"
	}
	return "1"
}

func handelsregisterPageSize(limit int) string {
	switch {
	case limit <= 10:
		return "10"
	case limit <= 25:
		return "25"
	case limit <= 50:
		return "50"
	default:
		return "100"
	}
}

func parseHTML(body []byte) (*bytes.Reader, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("empty html response")
	}
	return bytes.NewReader(body), nil
}
