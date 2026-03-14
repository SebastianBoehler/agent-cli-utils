package mdconv

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

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

func buildZip(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		handle, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := handle.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buffer.Bytes()
}
