package company

import (
	"strings"

	"golang.org/x/net/html"
)

func findNode(root *html.Node, predicate func(*html.Node) bool) *html.Node {
	if root == nil {
		return nil
	}
	if predicate(root) {
		return root
	}
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if match := findNode(child, predicate); match != nil {
			return match
		}
	}
	return nil
}

func findNodes(root *html.Node, predicate func(*html.Node) bool) []*html.Node {
	if root == nil {
		return nil
	}

	var matches []*html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if predicate(node) {
			matches = append(matches, node)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return matches
}

func attr(node *html.Node, key string) string {
	for _, item := range node.Attr {
		if item.Key == key {
			return item.Val
		}
	}
	return ""
}

func nodeText(node *html.Node) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
			builder.WriteByte(' ')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return cleanText(builder.String())
}

func cleanText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}
