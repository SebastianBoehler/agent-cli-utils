package company

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const (
	SourceAll               = "all"
	SourceHandelsregister   = "handelsregister"
	SourceOffeneRegister    = "offeneregister"
	SourceOpenCorporates    = "opencorporates"
	defaultCountry          = "de"
	defaultLimit            = 10
	maxLimit                = 100
	defaultTimeout          = 30 * time.Second
	defaultHandelsregister  = "https://www.handelsregister.de"
	defaultOffeneRegister   = "https://offeneregister.de/"
	defaultOffeneRegisterDB = "https://db.offeneregister.de"
	defaultOpenCorporates   = "https://api.opencorporates.com/v0.4"
	defaultUserAgent        = "company-cli/0.1"
)

var sourcePriority = []string{
	SourceHandelsregister,
	SourceOpenCorporates,
	SourceOffeneRegister,
}

type Config struct {
	HTTPClient             *http.Client
	Timeout                time.Duration
	UserAgent              string
	HandelsregisterURL     string
	HandelsregisterDelay   time.Duration
	OffeneRegisterURL      string
	OffeneRegisterDBURL    string
	OffeneRegisterSlug     string
	OpenCorporatesURL      string
	OpenCorporatesAPIToken string
}

type SearchRequest struct {
	Query   string `json:"query" yaml:"query"`
	City    string `json:"city,omitempty" yaml:"city,omitempty"`
	Country string `json:"country" yaml:"country"`
	Exact   bool   `json:"exact" yaml:"exact"`
	Limit   int    `json:"limit" yaml:"limit"`
}

type CompanyResult struct {
	Source            string         `json:"source" yaml:"source"`
	Name              string         `json:"name" yaml:"name"`
	RegisterNumber    string         `json:"register_number,omitempty" yaml:"register_number,omitempty"`
	Jurisdiction      string         `json:"jurisdiction,omitempty" yaml:"jurisdiction,omitempty"`
	Court             string         `json:"court,omitempty" yaml:"court,omitempty"`
	City              string         `json:"city,omitempty" yaml:"city,omitempty"`
	Status            string         `json:"status,omitempty" yaml:"status,omitempty"`
	CompanyType       string         `json:"company_type,omitempty" yaml:"company_type,omitempty"`
	Address           string         `json:"address,omitempty" yaml:"address,omitempty"`
	IncorporationDate string         `json:"incorporation_date,omitempty" yaml:"incorporation_date,omitempty"`
	DissolutionDate   string         `json:"dissolution_date,omitempty" yaml:"dissolution_date,omitempty"`
	RegistryURL       string         `json:"registry_url,omitempty" yaml:"registry_url,omitempty"`
	SourceURL         string         `json:"source_url,omitempty" yaml:"source_url,omitempty"`
	Raw               map[string]any `json:"raw,omitempty" yaml:"raw,omitempty"`
}

type SourceError struct {
	Source string `json:"source" yaml:"source"`
	Error  string `json:"error" yaml:"error"`
}

type SearchReport struct {
	Query   string          `json:"query" yaml:"query"`
	City    string          `json:"city,omitempty" yaml:"city,omitempty"`
	Country string          `json:"country" yaml:"country"`
	Exact   bool            `json:"exact" yaml:"exact"`
	Sources []string        `json:"sources" yaml:"sources"`
	Results []CompanyResult `json:"results" yaml:"results"`
	Errors  []SourceError   `json:"errors,omitempty" yaml:"errors,omitempty"`
}

func (report SearchReport) AllSourcesFailed() bool {
	return len(report.Sources) > 0 && len(report.Errors) == len(report.Sources)
}

type QuotaReport struct {
	Source    string `json:"source" yaml:"source"`
	SourceURL string `json:"source_url" yaml:"source_url"`
	Data      any    `json:"data,omitempty" yaml:"data,omitempty"`
}

type Provider interface {
	Name() string
	Search(ctx context.Context, req SearchRequest) ([]CompanyResult, error)
}

type QuotaProvider interface {
	Quota(ctx context.Context) (QuotaReport, error)
}

func normalizeRequest(req SearchRequest) SearchRequest {
	req.Query = strings.TrimSpace(req.Query)
	req.City = strings.TrimSpace(req.City)
	req.Country = strings.ToLower(strings.TrimSpace(req.Country))
	if req.Country == "" {
		req.Country = defaultCountry
	}
	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}
	return req
}
