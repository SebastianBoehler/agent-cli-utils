package mdconv

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestDetectFormatByContent(t *testing.T) {
	service := NewService()
	cases := []struct {
		name   string
		input  []byte
		format string
	}{
		{name: "json", input: []byte(`{"ok":true}`), format: "json"},
		{name: "xml", input: []byte(`<?xml version="1.0"?><root><item>1</item></root>`), format: "xml"},
		{name: "html", input: []byte(`<html><body><p>hi</p></body></html>`), format: "html"},
		{name: "csv", input: []byte("name,age\nsam,8\n"), format: "csv"},
		{name: "text", input: []byte("plain text body"), format: "txt"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := service.detectFormat("", tc.input); got != tc.format {
				t.Fatalf("detect format = %q, want %q", got, tc.format)
			}
		})
	}
}

func TestConvertHTML(t *testing.T) {
	service := NewService()
	result, err := service.Convert([]byte("<html><body><h1>Title</h1><p>Hello <strong>world</strong>.</p></body></html>"), Options{Name: "index.html"})
	if err != nil {
		t.Fatalf("convert html: %v", err)
	}

	if !strings.Contains(result.Markdown, "# Title") {
		t.Fatalf("expected heading, got %q", result.Markdown)
	}
	if !strings.Contains(result.Markdown, "Hello **world** .") && !strings.Contains(result.Markdown, "Hello **world**.") {
		t.Fatalf("expected paragraph markdown, got %q", result.Markdown)
	}
}

func TestConvertCSV(t *testing.T) {
	service := NewService()
	result, err := service.Convert([]byte("name,score\nalice,10\nbob,7\n"), Options{Name: "scores.csv"})
	if err != nil {
		t.Fatalf("convert csv: %v", err)
	}

	if !strings.Contains(result.Markdown, "| name | score |") {
		t.Fatalf("expected markdown table, got %q", result.Markdown)
	}
}

func TestConvertJSONAndYAMLAndXML(t *testing.T) {
	service := NewService()

	jsonResult, err := service.Convert([]byte(`{"name":"alice","score":10}`), Options{Name: "sample.json"})
	if err != nil {
		t.Fatalf("convert json: %v", err)
	}
	if !strings.Contains(jsonResult.Markdown, "```json") {
		t.Fatalf("expected json code fence, got %q", jsonResult.Markdown)
	}

	yamlResult, err := service.Convert([]byte("name: alice\nscore: 10\n"), Options{Name: "sample.yaml"})
	if err != nil {
		t.Fatalf("convert yaml: %v", err)
	}
	if !strings.Contains(yamlResult.Markdown, "```yaml") {
		t.Fatalf("expected yaml code fence, got %q", yamlResult.Markdown)
	}

	xmlResult, err := service.Convert([]byte(`<?xml version="1.0"?><root><item>10</item></root>`), Options{Name: "sample.xml"})
	if err != nil {
		t.Fatalf("convert xml: %v", err)
	}
	if !strings.Contains(xmlResult.Markdown, "```xml") {
		t.Fatalf("expected xml code fence, got %q", xmlResult.Markdown)
	}
}

func TestConvertDOCX(t *testing.T) {
	service := NewService()
	data := buildZip(t, map[string]string{
		"word/document.xml": `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>Document Title</w:t></w:r></w:p><w:p><w:r><w:t>First paragraph.</w:t></w:r></w:p><w:tbl><w:tr><w:tc><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>B</w:t></w:r></w:p></w:tc></w:tr><w:tr><w:tc><w:p><w:r><w:t>1</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>2</w:t></w:r></w:p></w:tc></w:tr></w:tbl></w:body></w:document>`,
		"word/styles.xml":   `<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:style w:styleId="Heading1"><w:name w:val="Heading 1"/></w:style></w:styles>`,
	})

	result, err := service.Convert(data, Options{Name: "sample.docx"})
	if err != nil {
		t.Fatalf("convert docx: %v", err)
	}

	if !strings.Contains(result.Markdown, "# Document Title") {
		t.Fatalf("expected heading markdown, got %q", result.Markdown)
	}
	if !strings.Contains(result.Markdown, "| A | B |") {
		t.Fatalf("expected table markdown, got %q", result.Markdown)
	}
}

func TestConvertXLSX(t *testing.T) {
	service := NewService()
	data := buildZip(t, map[string]string{
		"xl/workbook.xml":            `<workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Sheet1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/sharedStrings.xml":       `<sst><si><t>Name</t></si><si><t>Score</t></si><si><t>Alice</t></si></sst>`,
		"xl/worksheets/sheet1.xml":   `<worksheet><sheetData><row><c t="s"><v>0</v></c><c t="s"><v>1</v></c></row><row><c t="s"><v>2</v></c><c><v>10</v></c></row></sheetData></worksheet>`,
	})

	result, err := service.Convert(data, Options{Name: "sample.xlsx"})
	if err != nil {
		t.Fatalf("convert xlsx: %v", err)
	}

	if !strings.Contains(result.Markdown, "## Sheet1") {
		t.Fatalf("expected sheet heading, got %q", result.Markdown)
	}
	if !strings.Contains(result.Markdown, "| Alice | 10 |") {
		t.Fatalf("expected worksheet table, got %q", result.Markdown)
	}
}

func TestConvertPPTX(t *testing.T) {
	service := NewService()
	data := buildZip(t, map[string]string{
		"ppt/presentation.xml":  `<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"/>`,
		"ppt/slides/slide1.xml": `<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:cSld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"><p:spTree><p:sp><p:txBody><a:p><a:r><a:t>Slide title</a:t></a:r></a:p><a:p><a:r><a:t>Slide body</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`,
		"ppt/slides/slide2.xml": `<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:cSld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"><p:spTree><p:sp><p:txBody><a:p><a:r><a:t>Second slide</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`,
	})

	result, err := service.Convert(data, Options{Name: "deck.pptx"})
	if err != nil {
		t.Fatalf("convert pptx: %v", err)
	}

	if !strings.Contains(result.Markdown, "## Slide 1") {
		t.Fatalf("expected slide heading, got %q", result.Markdown)
	}
	if !strings.Contains(result.Markdown, "Slide body") {
		t.Fatalf("expected slide text, got %q", result.Markdown)
	}
}

func TestConvertZIP(t *testing.T) {
	service := NewService()
	data := buildZip(t, map[string]string{
		"notes.txt":       "hello from zip",
		"nested/data.csv": "name,age\nsam,8\n",
	})

	result, err := service.Convert(data, Options{Name: "bundle.zip"})
	if err != nil {
		t.Fatalf("convert zip: %v", err)
	}

	if !strings.Contains(result.Markdown, "## notes.txt") {
		t.Fatalf("expected archive heading, got %q", result.Markdown)
	}
	if !strings.Contains(result.Markdown, "| name | age |") {
		t.Fatalf("expected nested csv markdown, got %q", result.Markdown)
	}
}

func TestConvertNestedZIPRespectsDepth(t *testing.T) {
	service := NewService()

	innerZip := buildZip(t, map[string]string{
		"deep.txt": "nested payload",
	})
	outerZip := buildZipBytes(t, map[string][]byte{
		"top.txt":   []byte("top level"),
		"inner.zip": innerZip,
	})

	result, err := service.Convert(outerZip, Options{Name: "bundle.zip", ArchiveDepth: 1})
	if err != nil {
		t.Fatalf("convert outer zip: %v", err)
	}

	if len(result.Warnings) == 0 {
		t.Fatalf("expected nesting warning, got none")
	}
	if strings.Contains(result.Markdown, "nested payload") {
		t.Fatalf("expected nested archive content to be skipped at depth limit, got %q", result.Markdown)
	}
}

func TestConvertZIPLargeEntryWarning(t *testing.T) {
	service := NewService()

	largeText := strings.Repeat("a", maxArchiveEntryBytes+1)
	data := buildZip(t, map[string]string{
		"too-large.txt": largeText,
	})

	_, err := service.Convert(data, Options{Name: "bundle.zip"})
	if err == nil {
		t.Fatal("expected error for unsupported oversized archive-only payload")
	}
	if !strings.Contains(err.Error(), "no supported files found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestForceFormatOverridesDetection(t *testing.T) {
	service := NewService()

	result, err := service.Convert([]byte("<p>forced html</p>"), Options{Name: "notes.txt", Format: "html"})
	if err != nil {
		t.Fatalf("forced format convert: %v", err)
	}
	if !strings.Contains(result.Markdown, "forced html") {
		t.Fatalf("expected forced html conversion, got %q", result.Markdown)
	}
}

func TestUnsupportedPDFReturnsError(t *testing.T) {
	service := NewService()

	_, err := service.Convert([]byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n"), Options{Name: "sample.pdf"})
	if err == nil {
		t.Fatal("expected unsupported pdf error")
	}
	if !strings.Contains(err.Error(), `unsupported input format "pdf"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func buildZip(t *testing.T, files map[string]string) []byte {
	rawFiles := make(map[string][]byte, len(files))
	for name, content := range files {
		rawFiles[name] = []byte(content)
	}
	return buildZipBytes(t, rawFiles)
}

func buildZipBytes(t *testing.T, files map[string][]byte) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		handle, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := handle.Write(content); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buffer.Bytes()
}

func TestMarkdownTableEscapesPipes(t *testing.T) {
	table := markdownTable([][]string{{"name", "note"}, {"alice", "a|b"}})
	if !strings.Contains(table, "a\\|b") {
		t.Fatalf("expected pipe escaping, got %q", table)
	}
}

func ExampleService_Convert() {
	service := NewService()
	result, err := service.Convert([]byte("name,score\nalice,10\n"), Options{Name: "scores.csv"})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(strings.Contains(result.Markdown, "| alice | 10 |"))
	// Output:
	// true
}
