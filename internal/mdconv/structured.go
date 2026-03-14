package mdconv

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

func convertPlainText(data []byte, opts convertOptions) (Result, error) {
	return Result{
		Source:   opts.Name,
		Format:   opts.Format,
		Markdown: normalizeNewlines(string(data)),
	}, nil
}

func convertCSV(data []byte, opts convertOptions) (Result, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	rows, err := reader.ReadAll()
	if err != nil {
		return Result{}, fmt.Errorf("parse csv: %w", err)
	}

	return Result{
		Source:   opts.Name,
		Format:   "csv",
		Markdown: markdownTable(rows),
	}, nil
}

func convertJSON(data []byte, opts convertOptions) (Result, error) {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return Result{}, fmt.Errorf("parse json: %w", err)
	}

	formatted, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("render json: %w", err)
	}

	return Result{
		Source:   opts.Name,
		Format:   "json",
		Markdown: markdownCodeFence("json", string(formatted)),
	}, nil
}

func convertYAML(data []byte, opts convertOptions) (Result, error) {
	var value any
	if err := yaml.Unmarshal(data, &value); err != nil {
		return Result{}, fmt.Errorf("parse yaml: %w", err)
	}

	formatted, err := yaml.Marshal(value)
	if err != nil {
		return Result{}, fmt.Errorf("render yaml: %w", err)
	}

	return Result{
		Source:   opts.Name,
		Format:   "yaml",
		Markdown: markdownCodeFence("yaml", string(formatted)),
	}, nil
}

func convertXML(data []byte, opts convertOptions) (Result, error) {
	if !xmlWellFormed(data) {
		return Result{}, fmt.Errorf("parse xml: invalid xml document")
	}

	formatted, err := indentXML(data)
	if err != nil {
		return Result{}, fmt.Errorf("render xml: %w", err)
	}

	return Result{
		Source:   opts.Name,
		Format:   "xml",
		Markdown: markdownCodeFence("xml", string(formatted)),
	}, nil
}

func xmlWellFormed(data []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		if _, err := decoder.Token(); err != nil {
			return err == io.EOF
		}
	}
}

func indentXML(data []byte) ([]byte, error) {
	var out bytes.Buffer
	decoder := xml.NewDecoder(bytes.NewReader(data))
	encoder := xml.NewEncoder(&out)
	encoder.Indent("", "  ")
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if err := encoder.EncodeToken(token); err != nil {
			return nil, err
		}
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
