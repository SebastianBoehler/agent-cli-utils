package company

import (
	"fmt"
	"strings"

	"github.com/SebastianBoehler/agent-cli-utils/internal/datax"
)

func RenderSearchText(report SearchReport) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "query: %s\n", report.Query)
	fmt.Fprintf(&builder, "country: %s\n", report.Country)
	if report.City != "" {
		fmt.Fprintf(&builder, "city: %s\n", report.City)
	}
	fmt.Fprintf(&builder, "exact: %t\n", report.Exact)
	fmt.Fprintf(&builder, "sources: %s\n", strings.Join(report.Sources, ", "))
	fmt.Fprintf(&builder, "results: %d\n", len(report.Results))
	if len(report.Errors) > 0 {
		fmt.Fprintf(&builder, "source_errors: %d\n", len(report.Errors))
	}

	for index, result := range report.Results {
		fmt.Fprintf(&builder, "\n[%d] %s\n", index+1, result.Name)
		fmt.Fprintf(&builder, "source: %s\n", result.Source)
		writeOptional(&builder, "register_number", result.RegisterNumber)
		writeOptional(&builder, "jurisdiction", result.Jurisdiction)
		writeOptional(&builder, "court", result.Court)
		writeOptional(&builder, "city", result.City)
		writeOptional(&builder, "status", result.Status)
		writeOptional(&builder, "company_type", result.CompanyType)
		writeOptional(&builder, "address", result.Address)
		writeOptional(&builder, "incorporation_date", result.IncorporationDate)
		writeOptional(&builder, "dissolution_date", result.DissolutionDate)
		writeOptional(&builder, "registry_url", result.RegistryURL)
		writeOptional(&builder, "source_url", result.SourceURL)
	}

	if len(report.Errors) > 0 {
		builder.WriteString("\nerrors:\n")
		for _, item := range report.Errors {
			fmt.Fprintf(&builder, "- %s: %s\n", item.Source, item.Error)
		}
	}

	return builder.String()
}

func RenderQuotaText(report QuotaReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "source: %s\n", report.Source)
	fmt.Fprintf(&builder, "source_url: %s\n", report.SourceURL)
	if report.Data != nil {
		builder.WriteString("data:\n")
		rendered, err := datax.Render(report.Data, "json")
		if err != nil {
			fmt.Fprintf(&builder, "%v\n", report.Data)
		} else {
			builder.Write(rendered)
		}
	}
	return builder.String()
}

func writeOptional(builder *strings.Builder, key string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	fmt.Fprintf(builder, "%s: %s\n", key, value)
}
