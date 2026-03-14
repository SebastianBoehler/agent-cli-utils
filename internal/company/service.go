package company

import (
	"context"
	"fmt"
	"strings"
)

type Service struct {
	providers     map[string]Provider
	quotaProvider QuotaProvider
}

func NewService(cfg Config) *Service {
	opencorporates := NewOpenCorporatesProvider(cfg)
	return &Service{
		providers: map[string]Provider{
			SourceHandelsregister: NewHandelsregisterProvider(cfg),
			SourceOffeneRegister:  NewOffeneRegisterProvider(cfg),
			SourceOpenCorporates:  opencorporates,
		},
		quotaProvider: opencorporates,
	}
}

func (service *Service) Search(ctx context.Context, selection string, req SearchRequest) (SearchReport, error) {
	sources, err := expandSources(selection)
	if err != nil {
		return SearchReport{}, err
	}

	req = normalizeRequest(req)
	report := SearchReport{
		Query:   req.Query,
		City:    req.City,
		Country: req.Country,
		Exact:   req.Exact,
		Sources: sources,
		Results: make([]CompanyResult, 0, req.Limit),
	}

	for _, source := range sources {
		provider := service.providers[source]
		if provider == nil {
			report.Errors = append(report.Errors, SourceError{Source: source, Error: "provider not configured"})
			continue
		}

		results, err := provider.Search(ctx, req)
		if err != nil {
			report.Errors = append(report.Errors, SourceError{Source: source, Error: err.Error()})
			continue
		}

		if req.Exact && source != SourceHandelsregister {
			results = filterExact(results, req.Query)
		}

		report.Results = append(report.Results, results...)
	}

	if len(report.Results) > req.Limit {
		report.Results = report.Results[:req.Limit]
	}

	return report, nil
}

func (service *Service) Quota(ctx context.Context) (QuotaReport, error) {
	if service.quotaProvider == nil {
		return QuotaReport{}, fmt.Errorf("quota provider not configured")
	}
	return service.quotaProvider.Quota(ctx)
}

func expandSources(selection string) ([]string, error) {
	source := strings.ToLower(strings.TrimSpace(selection))
	if source == "" || source == SourceAll {
		return append([]string(nil), sourcePriority...), nil
	}

	for _, known := range sourcePriority {
		if source == known {
			return []string{known}, nil
		}
	}

	return nil, fmt.Errorf("unsupported source %q", selection)
}

func filterExact(results []CompanyResult, query string) []CompanyResult {
	filtered := make([]CompanyResult, 0, len(results))
	for _, result := range results {
		if strings.EqualFold(strings.TrimSpace(result.Name), strings.TrimSpace(query)) {
			filtered = append(filtered, result)
		}
	}
	return filtered
}
