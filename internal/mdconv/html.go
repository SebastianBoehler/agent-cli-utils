package mdconv

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

func convertHTML(data []byte, opts convertOptions) (Result, error) {
	root, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return Result{}, fmt.Errorf("parse html: %w", err)
	}

	var builder strings.Builder
	renderHTMLNode(&builder, root, 0)
	return Result{
		Source:   opts.Name,
		Format:   "html",
		Markdown: normalizeNewlines(builder.String()),
	}, nil
}

func renderHTMLNode(builder *strings.Builder, node *html.Node, listDepth int) {
	if node == nil {
		return
	}

	switch node.Type {
	case html.TextNode:
		text := cleanText(node.Data)
		if text != "" {
			appendInline(builder, text)
		}
		return
	case html.ElementNode:
		renderHTMLElement(builder, node, listDepth)
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		renderHTMLNode(builder, child, listDepth)
	}
}

func renderHTMLElement(builder *strings.Builder, node *html.Node, listDepth int) {
	switch node.Data {
	case "script", "style", "noscript":
		return
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(node.Data[1] - '0')
		builder.WriteString(strings.Repeat("#", level))
		builder.WriteString(" ")
		builder.WriteString(extractInlineText(node))
		builder.WriteString("\n\n")
	case "p", "section", "article", "div":
		renderChildren(builder, node, listDepth)
		builder.WriteString("\n\n")
	case "br":
		builder.WriteString("\n")
	case "hr":
		builder.WriteString("\n---\n\n")
	case "strong", "b":
		writeWrappedInline(builder, "**", node)
	case "em", "i":
		writeWrappedInline(builder, "_", node)
	case "code":
		if node.Parent != nil && node.Parent.Data == "pre" {
			return
		}
		writeWrappedInline(builder, "`", node)
	case "pre":
		builder.WriteString(markdownCodeFence("", extractPreText(node)))
		builder.WriteString("\n\n")
	case "blockquote":
		lines := strings.Split(normalizeNewlines(extractInlineText(node)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			builder.WriteString("> ")
			builder.WriteString(line)
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	case "a":
		text := extractInlineText(node)
		href := findAttribute(node, "href")
		if href == "" {
			appendInline(builder, text)
			return
		}
		if text == "" {
			text = href
		}
		appendInline(builder, "["+text+"]("+href+")")
	case "ul":
		renderList(builder, node, listDepth, false)
		builder.WriteString("\n")
	case "ol":
		renderList(builder, node, listDepth, true)
		builder.WriteString("\n")
	case "table":
		table := extractHTMLTable(node)
		if len(table) > 0 {
			builder.WriteString(markdownTable(table))
			builder.WriteString("\n\n")
		}
	default:
		renderChildren(builder, node, listDepth)
	}
}

func renderChildren(builder *strings.Builder, node *html.Node, listDepth int) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		renderHTMLNode(builder, child, listDepth)
	}
}

func renderList(builder *strings.Builder, node *html.Node, listDepth int, ordered bool) {
	index := 1
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode || child.Data != "li" {
			continue
		}

		indent := strings.Repeat("  ", listDepth)
		marker := "- "
		if ordered {
			marker = fmt.Sprintf("%d. ", index)
		}

		builder.WriteString(indent)
		builder.WriteString(marker)
		renderListItem(builder, child, listDepth)
		builder.WriteString("\n")
		index++
	}
}

func renderListItem(builder *strings.Builder, node *html.Node, listDepth int) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && (child.Data == "ul" || child.Data == "ol") {
			builder.WriteString("\n")
			renderList(builder, child, listDepth+1, child.Data == "ol")
			continue
		}
		renderHTMLNode(builder, child, listDepth+1)
	}
}

func extractHTMLTable(node *html.Node) [][]string {
	var rows [][]string
	for row := node.FirstChild; row != nil; row = row.NextSibling {
		rows = appendTableRows(rows, row)
	}
	return rows
}

func appendTableRows(rows [][]string, node *html.Node) [][]string {
	if node.Type == html.ElementNode && node.Data == "tr" {
		var row []string
		for cell := node.FirstChild; cell != nil; cell = cell.NextSibling {
			if cell.Type == html.ElementNode && (cell.Data == "th" || cell.Data == "td") {
				row = append(row, extractInlineText(cell))
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		rows = appendTableRows(rows, child)
	}
	return rows
}

func extractInlineText(node *html.Node) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			text := cleanText(current.Data)
			if text != "" {
				parts = append(parts, text)
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(parts, " ")
}

func extractPreText(node *html.Node) string {
	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return normalizeNewlines(builder.String())
}

func writeWrappedInline(builder *strings.Builder, wrapper string, node *html.Node) {
	text := extractInlineText(node)
	if text == "" {
		return
	}
	appendInline(builder, wrapper+text+wrapper)
}

func findAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func appendInline(builder *strings.Builder, value string) {
	if value == "" {
		return
	}
	current := builder.String()
	if needsInlineSpace(current, value) {
		builder.WriteString(" ")
	}
	builder.WriteString(value)
}

func needsInlineSpace(current string, next string) bool {
	current = strings.TrimRight(current, " \n\t")
	next = strings.TrimLeft(next, " \n\t")
	if current == "" || next == "" {
		return false
	}

	last, _ := utf8.DecodeLastRuneInString(current)
	first, _ := utf8.DecodeRuneInString(next)
	if isInlinePunctuation(last) || isInlinePunctuation(first) {
		return false
	}
	return true
}

func isInlinePunctuation(value rune) bool {
	switch value {
	case '.', ',', ';', ':', '!', '?', ')', ']', '}', '>', '\n', ' ':
		return true
	case '(', '[', '{', '<':
		return true
	}
	return false
}
