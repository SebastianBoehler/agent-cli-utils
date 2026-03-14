package company

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type OpenCorporatesProvider struct {
	client    *http.Client
	baseURL   string
	apiToken  string
	userAgent string
}

func NewOpenCorporatesProvider(cfg Config) *OpenCorporatesProvider {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.OpenCorporatesURL), "/")
	if baseURL == "" {
		baseURL = defaultOpenCorporates
	}

	userAgent := strings.TrimSpace(cfg.UserAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	return &OpenCorporatesProvider{
		client:    cloneHTTPClient(cfg.HTTPClient, cfg.Timeout),
		baseURL:   baseURL,
		apiToken:  strings.TrimSpace(cfg.OpenCorporatesAPIToken),
		userAgent: userAgent,
	}
}

func (provider *OpenCorporatesProvider) Name() string {
	return SourceOpenCorporates
}

func (provider *OpenCorporatesProvider) Search(ctx context.Context, req SearchRequest) ([]CompanyResult, error) {
	if provider.apiToken == "" {
		return nil, fmt.Errorf("missing OpenCorporates API token")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.baseURL+"/companies/search", nil)
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}
	request.Header.Set("User-Agent", provider.userAgent)

	query := request.URL.Query()
	query.Set("api_token", provider.apiToken)
	query.Set("q", req.Query)
	query.Set("country_code", req.Country)
	query.Set("order", "score")
	query.Set("per_page", fmt.Sprintf("%d", req.Limit))
	query.Set("normalise_company_name", "true")
	request.URL.RawQuery = query.Encode()

	response, err := provider.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	body, err := readResponse(response)
	if err != nil {
		return nil, err
	}

	return parseOpenCorporatesResults(body)
}

func (provider *OpenCorporatesProvider) Quota(ctx context.Context) (QuotaReport, error) {
	if provider.apiToken == "" {
		return QuotaReport{}, fmt.Errorf("missing OpenCorporates API token")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.baseURL+"/account_status", nil)
	if err != nil {
		return QuotaReport{}, fmt.Errorf("build quota request: %w", err)
	}
	request.Header.Set("User-Agent", provider.userAgent)

	query := request.URL.Query()
	query.Set("api_token", provider.apiToken)
	request.URL.RawQuery = query.Encode()

	response, err := provider.client.Do(request)
	if err != nil {
		return QuotaReport{}, fmt.Errorf("quota request failed: %w", err)
	}

	body, err := readResponse(response)
	if err != nil {
		return QuotaReport{}, err
	}

	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return QuotaReport{}, fmt.Errorf("parse quota response: %w", err)
	}

	return QuotaReport{
		Source:    SourceOpenCorporates,
		SourceURL: provider.baseURL,
		Data:      payload,
	}, nil
}

func parseOpenCorporatesResults(body []byte) ([]CompanyResult, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse opencorporates response: %w", err)
	}

	results, ok := payload["results"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("parse opencorporates response: missing results")
	}

	items, ok := results["companies"].([]any)
	if !ok {
		return nil, fmt.Errorf("parse opencorporates response: missing companies")
	}

	companies := make([]CompanyResult, 0, len(items))
	for _, item := range items {
		wrapper, ok := item.(map[string]any)
		if !ok {
			continue
		}
		company, ok := wrapper["company"].(map[string]any)
		if !ok {
			continue
		}

		companies = append(companies, CompanyResult{
			Source:            SourceOpenCorporates,
			Name:              stringValue(company["name"]),
			RegisterNumber:    stringValue(company["company_number"]),
			Jurisdiction:      stringValue(company["jurisdiction_code"]),
			City:              opencorporatesCity(company),
			Status:            stringValue(company["current_status"]),
			CompanyType:       stringValue(company["company_type"]),
			Address:           firstNonEmpty(stringValue(company["registered_address_in_full"]), nestedString(company, "registered_address", "street_address")),
			IncorporationDate: stringValue(company["incorporation_date"]),
			DissolutionDate:   stringValue(company["dissolution_date"]),
			RegistryURL:       stringValue(company["registry_url"]),
			SourceURL:         stringValue(company["opencorporates_url"]),
			Raw:               company,
		})
	}
	return companies, nil
}

func opencorporatesCity(company map[string]any) string {
	return firstNonEmpty(
		nestedString(company, "registered_address", "locality"),
		nestedString(company, "registered_address", "city"),
		nestedString(company, "registered_address", "locality_name"),
	)
}

func nestedString(object map[string]any, keys ...string) string {
	current := any(object)
	for _, key := range keys {
		node, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = node[key]
	}
	return stringValue(current)
}
