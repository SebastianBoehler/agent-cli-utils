package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/SebastianBoehler/agent-cli-utils/internal/company"
)

func TestRunSearchTextOutput(t *testing.T) {
	var stdout bytes.Buffer
	exitCode, err := run([]string{"search", "--format", "text", "Acme GmbH"}, &stdout, func(string) string { return "" }, func(cfg company.Config) service {
		return stubService{
			searchReport: company.SearchReport{
				Query:   "Acme GmbH",
				Country: "de",
				Sources: []string{company.SourceHandelsregister},
				Results: []company.CompanyResult{{
					Source:         company.SourceHandelsregister,
					Name:           "Acme GmbH",
					RegisterNumber: "HRB 123",
				}},
			},
		}
	})
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout.String(), "Acme GmbH") {
		t.Fatalf("stdout = %q, want company name", stdout.String())
	}
}

func TestRunSearchReturnsNonZeroWhenAllSourcesFail(t *testing.T) {
	var stdout bytes.Buffer
	exitCode, err := run([]string{"search", "Acme"}, &stdout, func(string) string { return "" }, func(cfg company.Config) service {
		return stubService{
			searchReport: company.SearchReport{
				Query:   "Acme",
				Country: "de",
				Sources: []string{company.SourceHandelsregister, company.SourceOpenCorporates},
				Errors: []company.SourceError{
					{Source: company.SourceHandelsregister, Error: "bad gateway"},
					{Source: company.SourceOpenCorporates, Error: "token missing"},
				},
			},
		}
	})
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}
}

func TestRunQuotaWritesJSON(t *testing.T) {
	var stdout bytes.Buffer
	exitCode, err := run([]string{"quota"}, &stdout, func(key string) string {
		if key == "OPENCORPORATES_API_TOKEN" {
			return "secret"
		}
		return ""
	}, func(cfg company.Config) service {
		if cfg.OpenCorporatesAPIToken != "secret" {
			t.Fatalf("OpenCorporatesAPIToken = %q, want secret", cfg.OpenCorporatesAPIToken)
		}
		return stubService{
			quotaReport: company.QuotaReport{
				Source:    company.SourceOpenCorporates,
				SourceURL: "https://api.opencorporates.com/v0.4",
				Data: map[string]any{
					"results": map[string]any{"account_status": map[string]any{"plan": "starter"}},
				},
			},
		}
	})
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout.String(), `"source": "opencorporates"`) {
		t.Fatalf("stdout = %q, want quota json", stdout.String())
	}
}

type stubService struct {
	searchReport company.SearchReport
	searchErr    error
	quotaReport  company.QuotaReport
	quotaErr     error
}

func (service stubService) Search(context.Context, string, company.SearchRequest) (company.SearchReport, error) {
	return service.searchReport, service.searchErr
}

func (service stubService) Quota(context.Context) (company.QuotaReport, error) {
	return service.quotaReport, service.quotaErr
}
