package mdconv

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
)

const maxArchiveEntryBytes = 8 << 20

func (service *Service) convertZIP(data []byte, opts convertOptions) (Result, error) {
	if opts.ArchiveDepth <= 0 {
		return Result{}, fmt.Errorf("archive nesting limit exceeded")
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return Result{}, fmt.Errorf("open zip: %w", err)
	}

	files := make([]*zip.File, 0, len(reader.File))
	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			files = append(files, file)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	if len(files) > opts.MaxArchiveEntries {
		files = files[:opts.MaxArchiveEntries]
	}

	var parts []string
	var warnings []string
	for _, file := range files {
		payload, err := readZipFile(file)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", file.Name, err))
			continue
		}

		childOpts := opts
		childOpts.Name = file.Name
		childOpts.Format = ""
		childOpts.ArchiveDepth--

		result, err := service.Convert(payload, Options{
			Name:              childOpts.Name,
			MaxArchiveEntries: childOpts.MaxArchiveEntries,
			ArchiveDepth:      childOpts.ArchiveDepth,
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", file.Name, err))
			continue
		}

		parts = append(parts, renderArchiveEntry(file.Name, result.Markdown))
	}

	if len(parts) == 0 && len(warnings) > 0 {
		return Result{}, fmt.Errorf("no supported files found in archive")
	}

	return Result{
		Source:   opts.Name,
		Format:   "zip",
		Markdown: normalizeNewlines(strings.Join(parts, "\n\n")),
		Warnings: warnings,
	}, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	if file.UncompressedSize64 > maxArchiveEntryBytes {
		return nil, fmt.Errorf("entry exceeds %d bytes", maxArchiveEntryBytes)
	}

	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func renderArchiveEntry(name string, markdown string) string {
	var builder strings.Builder
	builder.WriteString("## ")
	builder.WriteString(path.Clean(name))
	builder.WriteString("\n\n")
	builder.WriteString(strings.TrimSpace(markdown))
	return builder.String()
}
