package company

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServiceSearchKeepsPerSourceErrors(t *testing.T) {
	service := &Service{
		providers: map[string]Provider{
			SourceHandelsregister: fakeProvider{
				source: SourceHandelsregister,
				results: []CompanyResult{{
					Source: SourceHandelsregister,
					Name:   "Acme GmbH",
				}},
			},
			SourceOpenCorporates: fakeProvider{
				source: SourceOpenCorporates,
				err:    errString("token missing"),
			},
			SourceOffeneRegister: fakeProvider{
				source: SourceOffeneRegister,
			},
		},
	}

	report, err := service.Search(context.Background(), SourceAll, SearchRequest{Query: "Acme", Country: "de", Limit: 10})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(report.Results) != 1 {
		t.Fatalf("len(report.Results) = %d, want 1", len(report.Results))
	}
	if len(report.Errors) != 1 || report.Errors[0].Source != SourceOpenCorporates {
		t.Fatalf("report.Errors = %#v, want opencorporates error", report.Errors)
	}
	if report.AllSourcesFailed() {
		t.Fatal("AllSourcesFailed() = true, want false")
	}
}

func TestHandelsregisterProviderSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			_, _ = writer.Write([]byte(`<html><body><form><input name="javax.faces.ViewState" value="state-1"></form></body></html>`))
		case http.MethodPost:
			if got := request.Header.Get("Content-Type"); !strings.Contains(got, "application/x-www-form-urlencoded") {
				t.Fatalf("Content-Type = %q, want form encoded", got)
			}
			if err := request.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			if request.Form.Get("form:schlagwoerter") != "Acme" {
				t.Fatalf("query = %q, want Acme", request.Form.Get("form:schlagwoerter"))
			}
			_, _ = writer.Write([]byte(`<html><body><table role="grid"><tbody><tr data-ri="0"><td>1</td><td>Amtsgericht Berlin HRB 12345 B</td><td><a href="/detail/1">Acme GmbH</a></td><td>Berlin</td><td>Aktiv</td></tr></tbody></table></body></html>`))
		default:
			t.Fatalf("method = %s, want GET or POST", request.Method)
		}
	}))
	defer server.Close()

	provider := NewHandelsregisterProvider(Config{
		HTTPClient:           server.Client(),
		HandelsregisterURL:   server.URL,
		HandelsregisterDelay: 0,
	})

	results, err := provider.Search(context.Background(), SearchRequest{Query: "Acme", Country: "de", Limit: 10})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].RegisterNumber != "HRB 12345 B" {
		t.Fatalf("RegisterNumber = %q, want HRB 12345 B", results[0].RegisterNumber)
	}
	if results[0].Court != "Amtsgericht Berlin" {
		t.Fatalf("Court = %q, want Amtsgericht Berlin", results[0].Court)
	}
}

func TestOffeneRegisterProviderSearch(t *testing.T) {
	dbServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/openregister-test.json" {
			t.Fatalf("path = %s, want /openregister-test.json", request.URL.Path)
		}
		_, _ = writer.Write([]byte(`[{"name":"Acme GmbH","company_number":"HRB 123","jurisdiction_code":"de","current_status":"Active"}]`))
	}))
	defer dbServer.Close()

	siteServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`<html><body><a href="` + dbServer.URL + `/openregister-test">db</a></body></html>`))
	}))
	defer siteServer.Close()

	provider := NewOffeneRegisterProvider(Config{
		HTTPClient:          siteServer.Client(),
		OffeneRegisterURL:   siteServer.URL,
		OffeneRegisterDBURL: dbServer.URL,
	})

	results, err := provider.Search(context.Background(), SearchRequest{Query: "Acme", Country: "de", Limit: 5})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].SourceURL != siteServer.URL {
		t.Fatalf("SourceURL = %q, want %q", results[0].SourceURL, siteServer.URL)
	}
}

func TestOpenCorporatesProviderSearchAndQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/companies/search":
			if request.URL.Query().Get("api_token") != "secret" {
				t.Fatalf("api_token = %q, want secret", request.URL.Query().Get("api_token"))
			}
			_, _ = writer.Write([]byte(`{"results":{"companies":[{"company":{"name":"Acme GmbH","company_number":"HRB 123","jurisdiction_code":"de","current_status":"Active","company_type":"gmbh","opencorporates_url":"https://example.com/company","registry_url":"https://registry.example.com/company"}}]}}`))
		case "/account_status":
			_, _ = writer.Write([]byte(`{"results":{"account_status":{"plan":"starter"}}}`))
		default:
			t.Fatalf("path = %s", request.URL.Path)
		}
	}))
	defer server.Close()

	provider := NewOpenCorporatesProvider(Config{
		HTTPClient:             server.Client(),
		OpenCorporatesURL:      server.URL,
		OpenCorporatesAPIToken: "secret",
	})

	results, err := provider.Search(context.Background(), SearchRequest{Query: "Acme", Country: "de", Limit: 5})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].RegistryURL != "https://registry.example.com/company" {
		t.Fatalf("RegistryURL = %q", results[0].RegistryURL)
	}

	quota, err := provider.Quota(context.Background())
	if err != nil {
		t.Fatalf("Quota() error = %v", err)
	}
	if quota.Source != SourceOpenCorporates {
		t.Fatalf("quota.Source = %q, want %q", quota.Source, SourceOpenCorporates)
	}
}

type fakeProvider struct {
	source  string
	results []CompanyResult
	err     error
}

func (provider fakeProvider) Name() string {
	return provider.source
}

func (provider fakeProvider) Search(context.Context, SearchRequest) ([]CompanyResult, error) {
	if provider.err != nil {
		return nil, provider.err
	}
	return provider.results, nil
}

type errString string

func (err errString) Error() string {
	return string(err)
}
