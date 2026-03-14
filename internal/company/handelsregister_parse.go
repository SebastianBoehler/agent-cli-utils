package company

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var registerNumberPattern = regexp.MustCompile(`(HRA|HRB|GnR|VR|PR|GsR)\s*\d+(?:\s+[A-Z]{1,2})?`)

func extractHandelsregisterViewState(body []byte) (string, error) {
	reader, err := parseHTML(body)
	if err != nil {
		return "", err
	}

	root, err := html.Parse(reader)
	if err != nil {
		return "", fmt.Errorf("parse bootstrap html: %w", err)
	}

	node := findNode(root, func(item *html.Node) bool {
		return item.Type == html.ElementNode && item.Data == "input" && attr(item, "name") == "javax.faces.ViewState"
	})
	if node == nil {
		return "", fmt.Errorf("javax.faces.ViewState not found")
	}

	value := strings.TrimSpace(attr(node, "value"))
	if value == "" {
		return "", fmt.Errorf("javax.faces.ViewState missing value")
	}
	return value, nil
}

func parseHandelsregisterResults(body []byte, registryURL string, baseURL string) ([]CompanyResult, error) {
	reader, err := parseHTML(body)
	if err != nil {
		return nil, err
	}

	root, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("parse search html: %w", err)
	}

	rows := findNodes(root, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "tr" && attr(node, "data-ri") != ""
	})

	results := make([]CompanyResult, 0, len(rows))
	for _, row := range rows {
		cells := rowCells(row)
		if len(cells) < 5 {
			continue
		}

		courtField := nodeText(cells[1])
		name := nodeText(cells[2])
		city := nodeText(cells[3])
		status := nodeText(cells[4])
		court, registerNumber := splitCourtAndRegister(courtField)
		sourceURL := firstRowLink(row, baseURL)
		raw := map[string]any{
			"court_register": courtField,
			"name":           name,
			"city":           city,
			"status":         status,
		}
		if sourceURL != "" {
			raw["source_url"] = sourceURL
		}

		results = append(results, CompanyResult{
			Source:         SourceHandelsregister,
			Name:           name,
			RegisterNumber: registerNumber,
			Jurisdiction:   defaultCountry,
			Court:          court,
			City:           city,
			Status:         status,
			RegistryURL:    registryURL,
			SourceURL:      sourceURL,
			Raw:            raw,
		})
	}

	return results, nil
}

func rowCells(row *html.Node) []*html.Node {
	var cells []*html.Node
	for child := row.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "td" {
			cells = append(cells, child)
		}
	}
	return cells
}

func splitCourtAndRegister(value string) (string, string) {
	registerNumber := strings.TrimSpace(registerNumberPattern.FindString(value))
	if registerNumber == "" {
		return cleanText(value), ""
	}

	court := cleanText(strings.Replace(value, registerNumber, "", 1))
	return court, registerNumber
}

func firstRowLink(row *html.Node, baseURL string) string {
	link := findNode(row, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "a" && strings.TrimSpace(attr(node, "href")) != ""
	})
	if link == nil {
		return ""
	}
	return resolveURL(baseURL, attr(link, "href"))
}
