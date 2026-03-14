package mdconv

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

func convertXLSX(data []byte, opts convertOptions) (Result, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return Result{}, fmt.Errorf("open xlsx: %w", err)
	}

	sharedStrings, _ := xlsxSharedStrings(reader)
	sheets, err := xlsxSheets(reader, sharedStrings)
	if err != nil {
		return Result{}, err
	}

	var parts []string
	for _, sheet := range sheets {
		parts = append(parts, "## "+sheet.Name+"\n\n"+markdownTable(sheet.Rows))
	}

	return Result{
		Source:   opts.Name,
		Format:   "xlsx",
		Markdown: normalizeNewlines(strings.Join(parts, "\n\n")),
	}, nil
}

type workbookSheet struct {
	Name string
	Rows [][]string
}

func xlsxSheets(reader *zip.Reader, sharedStrings []string) ([]workbookSheet, error) {
	workbookXML, err := zipEntry(reader, "xl/workbook.xml")
	if err != nil {
		return nil, err
	}
	relsXML, err := zipEntry(reader, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}

	sheets, err := parseWorkbookSheets(workbookXML)
	if err != nil {
		return nil, err
	}
	targets, err := parseWorkbookRelationships(relsXML)
	if err != nil {
		return nil, err
	}

	out := make([]workbookSheet, 0, len(sheets))
	for _, sheet := range sheets {
		target := targets[sheet.RelID]
		if target == "" {
			continue
		}
		data, err := zipEntry(reader, path.Clean("xl/"+target))
		if err != nil {
			return nil, err
		}
		rows, err := parseWorksheetRows(data, sharedStrings)
		if err != nil {
			return nil, err
		}
		out = append(out, workbookSheet{Name: sheet.Name, Rows: rows})
	}
	return out, nil
}

type sheetRef struct {
	Name  string
	RelID string
}

func parseWorkbookSheets(data []byte) ([]sheetRef, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var sheets []sheetRef
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return sheets, nil
		}
		if err != nil {
			return nil, fmt.Errorf("parse workbook: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "sheet" {
			continue
		}
		sheets = append(sheets, sheetRef{
			Name:  attrValue(start.Attr, "name"),
			RelID: attrValue(start.Attr, "id"),
		})
	}
}

func parseWorkbookRelationships(data []byte) (map[string]string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	rels := map[string]string{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return rels, nil
		}
		if err != nil {
			return nil, fmt.Errorf("parse workbook relationships: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "Relationship" {
			continue
		}
		rels[attrValue(start.Attr, "Id")] = strings.TrimPrefix(attrValue(start.Attr, "Target"), "/")
	}
}

func xlsxSharedStrings(reader *zip.Reader) ([]string, error) {
	data, err := zipEntry(reader, "xl/sharedStrings.xml")
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	var values []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return values, nil
		}
		if err != nil {
			return nil, fmt.Errorf("parse shared strings: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "si" {
			continue
		}
		value, err := parseSharedString(decoder, start)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
}

func parseSharedString(decoder *xml.Decoder, start xml.StartElement) (string, error) {
	var parts []string
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local == "t" {
				var text string
				if err := decoder.DecodeElement(&text, &typed); err != nil {
					return "", err
				}
				parts = append(parts, text)
			}
		case xml.EndElement:
			if typed.Name.Local == start.Name.Local {
				return strings.Join(parts, ""), nil
			}
		}
	}
}

func parseWorksheetRows(data []byte, sharedStrings []string) ([][]string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var rows [][]string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return rows, nil
		}
		if err != nil {
			return nil, fmt.Errorf("parse worksheet: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "row" {
			continue
		}
		row, err := parseWorksheetRow(decoder, start, sharedStrings)
		if err != nil {
			return nil, err
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}
}

func parseWorksheetRow(decoder *xml.Decoder, start xml.StartElement, sharedStrings []string) ([]string, error) {
	var row []string
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local == "c" {
				cell, err := parseWorksheetCell(decoder, typed, sharedStrings)
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

func parseWorksheetCell(decoder *xml.Decoder, start xml.StartElement, sharedStrings []string) (string, error) {
	cellType := attrValue(start.Attr, "t")
	var value string
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local == "v" || typed.Name.Local == "t" {
				if err := decoder.DecodeElement(&value, &typed); err != nil {
					return "", err
				}
			}
		case xml.EndElement:
			if typed.Name.Local == start.Name.Local {
				if cellType == "s" {
					index, err := strconv.Atoi(strings.TrimSpace(value))
					if err == nil && index >= 0 && index < len(sharedStrings) {
						return sharedStrings[index], nil
					}
				}
				return strings.TrimSpace(value), nil
			}
		}
	}
}
