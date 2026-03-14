package company

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
)

var offeneRegisterSlugPattern = regexp.MustCompile(`openregister-[A-Za-z0-9_-]+`)

type OffeneRegisterProvider struct {
	client    *http.Client
	siteURL   string
	dbURL     string
	slug      string
	userAgent string
}

func NewOffeneRegisterProvider(cfg Config) *OffeneRegisterProvider {
	siteURL := strings.TrimSpace(cfg.OffeneRegisterURL)
	if siteURL == "" {
		siteURL = defaultOffeneRegister
	}

	dbURL := strings.TrimRight(strings.TrimSpace(cfg.OffeneRegisterDBURL), "/")
	if dbURL == "" {
		dbURL = defaultOffeneRegisterDB
	}

	userAgent := strings.TrimSpace(cfg.UserAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	return &OffeneRegisterProvider{
		client:    cloneHTTPClient(cfg.HTTPClient, cfg.Timeout),
		siteURL:   siteURL,
		dbURL:     dbURL,
		slug:      strings.TrimSpace(cfg.OffeneRegisterSlug),
		userAgent: userAgent,
	}
}

func (provider *OffeneRegisterProvider) Name() string {
	return SourceOffeneRegister
}

func (provider *OffeneRegisterProvider) Search(ctx context.Context, req SearchRequest) ([]CompanyResult, error) {
	slug, err := provider.resolveSlug(ctx)
	if err != nil {
		return nil, err
	}

	sql := `select name, company_number, native_company_number, jurisdiction_code, current_status, company_type, incorporation_date, dissolution_date, registered_address, opencorporates_url from company where lower(name) like lower('%' || :q || '%') limit ` + strconv.Itoa(req.Limit)
	endpoint := provider.dbURL + "/" + slug + ".json"

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build query request: %w", err)
	}
	request.Header.Set("User-Agent", provider.userAgent)

	query := request.URL.Query()
	query.Set("sql", sql)
	query.Set("_shape", "objects")
	query.Set("q", req.Query)
	request.URL.RawQuery = query.Encode()

	body, err := provider.doQuery(request, buildOffeneRegisterFallbackURL(provider.dbURL, slug, sql, req.Query))
	if err != nil {
		return nil, err
	}

	return parseOffeneRegisterResults(body)
}

func (provider *OffeneRegisterProvider) resolveSlug(ctx context.Context) (string, error) {
	if provider.slug != "" {
		return provider.slug, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.siteURL, nil)
	if err != nil {
		return "", fmt.Errorf("build slug discovery request: %w", err)
	}
	request.Header.Set("User-Agent", provider.userAgent)

	response, err := provider.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("slug discovery failed: %w", err)
	}

	body, err := readResponse(response)
	if err != nil {
		return "", err
	}

	slug := offeneRegisterSlugPattern.FindString(string(body))
	if slug == "" {
		return "", fmt.Errorf("datasette slug not found")
	}
	provider.slug = slug
	return provider.slug, nil
}

func (provider *OffeneRegisterProvider) doQuery(primary *http.Request, fallbackURL string) ([]byte, error) {
	response, err := provider.client.Do(primary)
	if err == nil {
		body, readErr := readResponse(response)
		if readErr == nil {
			return body, nil
		}
		err = readErr
	}

	fallback, buildErr := http.NewRequestWithContext(primary.Context(), http.MethodGet, fallbackURL, nil)
	if buildErr != nil {
		return nil, fmt.Errorf("build fallback query request: %w", buildErr)
	}
	fallback.Header.Set("User-Agent", provider.userAgent)

	response, fallbackErr := provider.client.Do(fallback)
	if fallbackErr != nil {
		return nil, fmt.Errorf("query request failed: %v; fallback failed: %w", err, fallbackErr)
	}
	body, readErr := readResponse(response)
	if readErr != nil {
		return nil, fmt.Errorf("query request failed: %v; fallback failed: %w", err, readErr)
	}
	return body, nil
}

func parseOffeneRegisterResults(body []byte) ([]CompanyResult, error) {
	rows, err := parseOffeneRegisterPayload(body)
	if err != nil {
		return nil, err
	}

	results := make([]CompanyResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, CompanyResult{
			Source:            SourceOffeneRegister,
			Name:              stringValue(row["name"]),
			RegisterNumber:    firstNonEmpty(stringValue(row["company_number"]), stringValue(row["native_company_number"])),
			Jurisdiction:      stringValue(row["jurisdiction_code"]),
			Status:            stringValue(row["current_status"]),
			CompanyType:       stringValue(row["company_type"]),
			Address:           stringValue(row["registered_address"]),
			IncorporationDate: stringValue(row["incorporation_date"]),
			DissolutionDate:   stringValue(row["dissolution_date"]),
			SourceURL:         defaultOffeneRegister,
			Raw:               row,
		})
	}
	return results, nil
}

func parseOffeneRegisterPayload(body []byte) ([]map[string]any, error) {
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, fmt.Errorf("parse offeneregister response: %w", err)
	}

	switch typed := value.(type) {
	case []any:
		return toObjectSlice(typed), nil
	case map[string]any:
		if rows, ok := typed["rows"].([]any); ok {
			return toObjectSlice(rows), nil
		}
		if rows, ok := typed["data"].([]any); ok {
			return toObjectSlice(rows), nil
		}
	}

	return nil, fmt.Errorf("parse offeneregister response: unexpected payload shape")
}

func toObjectSlice(items []any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		object, ok := item.(map[string]any)
		if ok {
			out = append(out, object)
		}
	}
	return out
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "<nil>" {
		return ""
	}
	return text
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" && trimmed != "<nil>" {
			return trimmed
		}
	}
	return ""
}

func buildOffeneRegisterFallbackURL(baseURL string, slug string, sql string, queryValue string) string {
	values := neturl.Values{}
	values.Set("sql", sql)
	values.Set("_shape", "objects")
	values.Set("_format", "json")
	values.Set("q", queryValue)
	return baseURL + "/" + slug + "?" + values.Encode()
}
