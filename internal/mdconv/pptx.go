package mdconv

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

func convertPPTX(data []byte, opts convertOptions) (Result, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return Result{}, fmt.Errorf("open pptx: %w", err)
	}

	slides := slideEntries(reader)
	if len(slides) == 0 {
		return Result{}, fmt.Errorf("missing slides")
	}

	var parts []string
	for index, name := range slides {
		data, err := zipEntry(reader, name)
		if err != nil {
			return Result{}, err
		}
		text, err := slideText(data)
		if err != nil {
			return Result{}, err
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("## Slide %d\n\n%s", index+1, normalizeNewlines(text)))
	}

	return Result{
		Source:   opts.Name,
		Format:   "pptx",
		Markdown: normalizeNewlines(strings.Join(parts, "\n\n")),
	}, nil
}

func slideEntries(reader *zip.Reader) []string {
	var slides []string
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			slides = append(slides, file.Name)
		}
	}
	sort.Strings(slides)
	return slides
}

func slideText(data []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var parts []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return strings.Join(parts, "\n"), nil
		}
		if err != nil {
			return "", fmt.Errorf("parse slide: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "t" {
			continue
		}
		var text string
		if err := decoder.DecodeElement(&text, &start); err != nil {
			return "", err
		}
		text = strings.TrimSpace(text)
		if text != "" {
			parts = append(parts, text)
		}
	}
}
