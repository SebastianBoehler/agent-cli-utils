package mdconv

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var extensionFormats = map[string]string{
	".txt":      "txt",
	".md":       "md",
	".markdown": "md",
	".html":     "html",
	".htm":      "html",
	".csv":      "csv",
	".json":     "json",
	".yaml":     "yaml",
	".yml":      "yaml",
	".xml":      "xml",
	".zip":      "zip",
	".docx":     "docx",
	".xlsx":     "xlsx",
	".pptx":     "pptx",
}

func detectByExtension(name string, _ []byte) (string, bool) {
	ext := strings.ToLower(filepath.Ext(name))
	format, ok := extensionFormats[ext]
	return format, ok
}

func detectOOXMLFormat(_ string, data []byte) (string, bool) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", false
	}

	hasWord := false
	hasWorkbook := false
	hasPresentation := false
	for _, file := range reader.File {
		switch file.Name {
		case "word/document.xml":
			hasWord = true
		case "xl/workbook.xml":
			hasWorkbook = true
		case "ppt/presentation.xml":
			hasPresentation = true
		}
	}

	switch {
	case hasWord:
		return "docx", true
	case hasWorkbook:
		return "xlsx", true
	case hasPresentation:
		return "pptx", true
	case len(reader.File) > 0:
		return "zip", true
	default:
		return "", false
	}
}

func detectByContent(_ string, data []byte) (string, bool) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "txt", true
	}

	lower := strings.ToLower(string(firstBytes(trimmed, 256)))
	if strings.HasPrefix(lower, "<!doctype html") || strings.HasPrefix(lower, "<html") {
		return "html", true
	}

	if strings.HasPrefix(lower, "<?xml") {
		return "xml", true
	}

	if json.Valid(trimmed) {
		return "json", true
	}

	if looksLikeCSV(trimmed) {
		return "csv", true
	}

	if utf8.Valid(trimmed) {
		if bytes.Contains(trimmed, []byte("<body")) || bytes.Contains(trimmed, []byte("<p")) || bytes.Contains(trimmed, []byte("<div")) {
			return "html", true
		}
		if bytes.HasPrefix(trimmed, []byte("<")) && bytes.Contains(trimmed, []byte(">")) {
			return "xml", true
		}
		return "txt", true
	}

	return "", false
}

func looksLikeCSV(data []byte) bool {
	lines := strings.Split(strings.TrimSpace(string(firstBytes(data, 1024))), "\n")
	if len(lines) < 2 {
		return false
	}

	firstCommaCount := strings.Count(lines[0], ",")
	if firstCommaCount == 0 {
		return false
	}

	for _, line := range lines[1:] {
		if strings.Count(line, ",") == firstCommaCount {
			return true
		}
	}

	return false
}

func firstBytes(data []byte, limit int) []byte {
	if len(data) <= limit {
		return data
	}

	return data[:limit]
}
