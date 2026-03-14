package mdconv

import (
	"html"
	"strings"
)

func normalizeNewlines(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	value = strings.Join(lines, "\n")
	for strings.Contains(value, "\n\n\n") {
		value = strings.ReplaceAll(value, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(value)
}

func markdownTable(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	width := 0
	for _, row := range rows {
		if len(row) > width {
			width = len(row)
		}
	}
	if width == 0 {
		return ""
	}

	normalized := make([][]string, 0, len(rows))
	for _, row := range rows {
		out := make([]string, width)
		for i := 0; i < width; i++ {
			if i < len(row) {
				out[i] = escapeTableCell(row[i])
			}
		}
		normalized = append(normalized, out)
	}

	var builder strings.Builder
	writeTableRow(&builder, normalized[0])
	builder.WriteString("\n|")
	for i := 0; i < width; i++ {
		builder.WriteString(" --- |")
	}
	for _, row := range normalized[1:] {
		builder.WriteString("\n")
		writeTableRow(&builder, row)
	}
	return builder.String()
}

func writeTableRow(builder *strings.Builder, row []string) {
	builder.WriteString("|")
	for _, cell := range row {
		builder.WriteString(" ")
		builder.WriteString(cell)
		builder.WriteString(" |")
	}
}

func escapeTableCell(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\n", "<br>"))
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}

func markdownCodeFence(language string, content string) string {
	var builder strings.Builder
	builder.WriteString("```")
	builder.WriteString(language)
	builder.WriteString("\n")
	builder.WriteString(strings.TrimSpace(content))
	builder.WriteString("\n```")
	return builder.String()
}

func cleanText(value string) string {
	value = html.UnescapeString(value)
	return strings.Join(strings.Fields(value), " ")
}
