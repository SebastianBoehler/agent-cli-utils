package samsungcontrol

import (
	"fmt"
	"strings"
)

func RenderPairText(result PairResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "protocol: %s\n", result.Protocol)
	fmt.Fprintf(&builder, "target: %s\n", result.Target)
	fmt.Fprintf(&builder, "ok: %t\n", result.OK)
	if result.Detail != "" {
		fmt.Fprintf(&builder, "detail: %s\n", result.Detail)
	}
	if result.Token != "" {
		fmt.Fprintf(&builder, "token: %s\n", result.Token)
	}
	return builder.String()
}

func RenderActionText(result ActionResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "operation: %s\n", result.Operation)
	fmt.Fprintf(&builder, "protocol: %s\n", result.Protocol)
	fmt.Fprintf(&builder, "target: %s\n", result.Target)
	fmt.Fprintf(&builder, "ok: %t\n", result.OK)
	if result.Detail != "" {
		fmt.Fprintf(&builder, "detail: %s\n", result.Detail)
	}
	return builder.String()
}
