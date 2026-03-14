package mdconv

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func convertDOCX(data []byte, opts convertOptions) (Result, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return Result{}, fmt.Errorf("open docx: %w", err)
	}

	styles, err := docxStyles(reader)
	if err != nil {
		return Result{}, err
	}

	documentXML, err := zipEntry(reader, "word/document.xml")
	if err != nil {
		return Result{}, err
	}

	markdown, err := docxMarkdown(documentXML, styles)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Source:   opts.Name,
		Format:   "docx",
		Markdown: markdown,
	}, nil
}

func docxMarkdown(data []byte, styles map[string]string) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var parts []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("parse docx xml: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		switch start.Name.Local {
		case "p":
			text, styleID, isList, err := parseDocxParagraph(decoder, start)
			if err != nil {
				return "", err
			}
			text = normalizeNewlines(text)
			if text == "" {
				continue
			}
			parts = append(parts, renderDocxParagraph(text, styleName(styles, styleID), isList))
		case "tbl":
			table, err := parseDocxTable(decoder, start)
			if err != nil {
				return "", err
			}
			if len(table) > 0 {
				parts = append(parts, markdownTable(table))
			}
		}
	}

	return normalizeNewlines(strings.Join(parts, "\n\n")), nil
}

func parseDocxParagraph(decoder *xml.Decoder, start xml.StartElement) (string, string, bool, error) {
	var builder strings.Builder
	var styleID string
	isList := false
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", "", false, err
		}

		switch typed := token.(type) {
		case xml.StartElement:
			switch typed.Name.Local {
			case "pStyle":
				styleID = attrValue(typed.Attr, "val")
			case "numPr":
				isList = true
			case "t":
				var text string
				if err := decoder.DecodeElement(&text, &typed); err != nil {
					return "", "", false, err
				}
				builder.WriteString(text)
			case "tab":
				builder.WriteString("\t")
			case "br", "cr":
				builder.WriteString("\n")
			}
		case xml.EndElement:
			if typed.Name.Local == start.Name.Local {
				return builder.String(), styleID, isList, nil
			}
		}
	}
}

func parseDocxTable(decoder *xml.Decoder, start xml.StartElement) ([][]string, error) {
	var rows [][]string
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local == "tr" {
				row, err := parseDocxRow(decoder, typed)
				if err != nil {
					return nil, err
				}
				if len(row) > 0 {
					rows = append(rows, row)
				}
			}
		case xml.EndElement:
			if typed.Name.Local == start.Name.Local {
				return rows, nil
			}
		}
	}
}

func parseDocxRow(decoder *xml.Decoder, start xml.StartElement) ([]string, error) {
	var row []string
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local == "tc" {
				cell, err := parseDocxCell(decoder, typed)
				if err != nil {
					return nil, err
				}
				row = append(row, cell)
			}
		case xml.EndElement:
			if typed.Name.Local == start.Name.Local {
				return row, nil
			}
		}
	}
}

func parseDocxCell(decoder *xml.Decoder, start xml.StartElement) (string, error) {
	var parts []string
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local == "p" {
				text, _, _, err := parseDocxParagraph(decoder, typed)
				if err != nil {
					return "", err
				}
				if trimmed := normalizeNewlines(text); trimmed != "" {
					parts = append(parts, trimmed)
				}
			}
		case xml.EndElement:
			if typed.Name.Local == start.Name.Local {
				return strings.Join(parts, "\n"), nil
			}
		}
	}
}

func renderDocxParagraph(text string, style string, isList bool) string {
	switch headingLevel(style) {
	case 1:
		return "# " + text
	case 2:
		return "## " + text
	case 3:
		return "### " + text
	case 4:
		return "#### " + text
	case 5:
		return "##### " + text
	case 6:
		return "###### " + text
	}

	if isList {
		return "- " + strings.ReplaceAll(text, "\n", " ")
	}
	return text
}
