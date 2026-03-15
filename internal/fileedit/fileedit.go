package fileedit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Request struct {
	Edits []Edit `json:"edits" yaml:"edits"`
}

type Edit struct {
	Path       string `json:"path" yaml:"path"`
	Action     string `json:"action" yaml:"action"`
	Old        string `json:"old,omitempty" yaml:"old,omitempty"`
	New        string `json:"new,omitempty" yaml:"new,omitempty"`
	Anchor     string `json:"anchor,omitempty" yaml:"anchor,omitempty"`
	ReplaceAll bool   `json:"replace_all,omitempty" yaml:"replace_all,omitempty"`
	StartLine  int    `json:"start_line,omitempty" yaml:"start_line,omitempty"`
	EndLine    int    `json:"end_line,omitempty" yaml:"end_line,omitempty"`
	Create     bool   `json:"create,omitempty" yaml:"create,omitempty"`
}

type Options struct {
	DryRun     bool
	FailOnNoop bool
}

type Report struct {
	DryRun       bool         `json:"dry_run" yaml:"dry_run"`
	Applied      int          `json:"applied" yaml:"applied"`
	Changed      int          `json:"changed" yaml:"changed"`
	Results      []EditResult `json:"results" yaml:"results"`
	ChangedPaths []string     `json:"changed_paths" yaml:"changed_paths"`
}

type EditResult struct {
	Path        string `json:"path" yaml:"path"`
	Action      string `json:"action" yaml:"action"`
	Changed     bool   `json:"changed" yaml:"changed"`
	Created     bool   `json:"created,omitempty" yaml:"created,omitempty"`
	Occurrences int    `json:"occurrences,omitempty" yaml:"occurrences,omitempty"`
	BeforeBytes int    `json:"before_bytes" yaml:"before_bytes"`
	AfterBytes  int    `json:"after_bytes" yaml:"after_bytes"`
	StartLine   int    `json:"start_line,omitempty" yaml:"start_line,omitempty"`
	EndLine     int    `json:"end_line,omitempty" yaml:"end_line,omitempty"`
}

type fileState struct {
	Path     string
	Content  string
	Mode     os.FileMode
	Existed  bool
	Original string
}

func LoadRequest(payload []byte) (Request, error) {
	trimmed := strings.TrimSpace(string(payload))
	if trimmed == "" {
		return Request{}, fmt.Errorf("spec is empty")
	}

	var request Request
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal(payload, &request); err != nil {
			return Request{}, fmt.Errorf("invalid JSON spec: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(payload, &request); err != nil {
			return Request{}, fmt.Errorf("invalid YAML spec: %w", err)
		}
	}

	if len(request.Edits) == 0 {
		return Request{}, fmt.Errorf("spec contains no edits")
	}

	return request, nil
}

func Apply(request Request, options Options) (Report, error) {
	report := Report{
		DryRun:       options.DryRun,
		Applied:      len(request.Edits),
		Results:      make([]EditResult, 0, len(request.Edits)),
		ChangedPaths: make([]string, 0, len(request.Edits)),
	}

	states := make(map[string]*fileState, len(request.Edits))
	seenChanged := make(map[string]struct{}, len(request.Edits))
	for index, edit := range request.Edits {
		result, state, err := stageOne(edit, options, states)
		if err != nil {
			return Report{}, fmt.Errorf("edit %d: %w", index+1, err)
		}

		report.Results = append(report.Results, result)
		if result.Changed {
			report.Changed++
			if _, exists := seenChanged[result.Path]; !exists {
				seenChanged[result.Path] = struct{}{}
				report.ChangedPaths = append(report.ChangedPaths, result.Path)
			}
		}
		states[result.Path] = state
	}

	if options.DryRun {
		return report, nil
	}

	for _, path := range report.ChangedPaths {
		state := states[path]
		if err := writeState(state); err != nil {
			return Report{}, err
		}
	}

	return report, nil
}

func RenderText(report Report) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "dry_run: %t\n", report.DryRun)
	fmt.Fprintf(&builder, "applied: %d\n", report.Applied)
	fmt.Fprintf(&builder, "changed: %d\n", report.Changed)
	for _, result := range report.Results {
		fmt.Fprintf(
			&builder,
			"%s action=%s changed=%t created=%t before=%d after=%d occurrences=%d\n",
			result.Path,
			result.Action,
			result.Changed,
			result.Created,
			result.BeforeBytes,
			result.AfterBytes,
			result.Occurrences,
		)
	}
	return builder.String()
}

func stageOne(edit Edit, options Options, states map[string]*fileState) (EditResult, *fileState, error) {
	path, action, err := validateEdit(edit)
	if err != nil {
		return EditResult{}, nil, err
	}

	state, err := loadState(path, states)
	if err != nil {
		return EditResult{}, nil, err
	}

	updated, occurrences, err := transform(state.Content, edit, state.Existed, options.FailOnNoop)
	if err != nil {
		return EditResult{}, nil, fmt.Errorf("%s: %w", path, err)
	}

	result := EditResult{
		Path:        path,
		Action:      action,
		Occurrences: occurrences,
		BeforeBytes: len(state.Content),
		AfterBytes:  len(updated),
		StartLine:   edit.StartLine,
		EndLine:     edit.EndLine,
		Created:     !state.Existed,
		Changed:     state.Content != updated || !state.Existed,
	}

	if !result.Changed {
		return result, state, nil
	}

	next := *state
	next.Content = updated
	next.Existed = true
	return result, &next, nil
}

func readFile(path string) ([]byte, os.FileMode, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, false, nil
		}
		return nil, 0, false, fmt.Errorf("%s: stat file: %w", path, err)
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, false, fmt.Errorf("%s: read file: %w", path, err)
	}

	return payload, info.Mode().Perm(), true, nil
}

func loadState(path string, states map[string]*fileState) (*fileState, error) {
	if state, ok := states[path]; ok {
		copy := *state
		return &copy, nil
	}

	payload, mode, existed, err := readFile(path)
	if err != nil {
		return nil, err
	}

	return &fileState{
		Path:     path,
		Content:  string(payload),
		Mode:     mode,
		Existed:  existed,
		Original: string(payload),
	}, nil
}

func writeState(state *fileState) error {
	if state == nil {
		return fmt.Errorf("missing file state")
	}
	if state.Content == state.Original && state.Existed {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(state.Path), 0o755); err != nil {
		return fmt.Errorf("%s: create parent directories: %w", state.Path, err)
	}

	mode := state.Mode
	if mode == 0 {
		mode = 0o644
	}

	if err := os.WriteFile(state.Path, []byte(state.Content), mode); err != nil {
		return fmt.Errorf("%s: write file: %w", state.Path, err)
	}

	return nil
}

func validateEdit(edit Edit) (string, string, error) {
	if strings.TrimSpace(edit.Path) == "" {
		return "", "", fmt.Errorf("edit path is required")
	}
	if strings.TrimSpace(edit.Action) == "" {
		return "", "", fmt.Errorf("edit action is required for %s", edit.Path)
	}

	path := filepath.Clean(edit.Path)
	action := strings.ToLower(strings.TrimSpace(edit.Action))

	switch action {
	case "replace":
		if edit.Old == "" {
			return "", "", fmt.Errorf("%s: replace requires old", path)
		}
	case "insert_before", "insert_after":
		if edit.Anchor == "" {
			return "", "", fmt.Errorf("%s: %s requires anchor", path, action)
		}
	case "replace_lines":
		if edit.StartLine <= 0 || edit.EndLine <= 0 || edit.EndLine < edit.StartLine {
			return "", "", fmt.Errorf("%s: replace_lines requires valid start_line and end_line", path)
		}
	case "append", "write":
	default:
		return "", "", fmt.Errorf("%s: unsupported action %q", path, edit.Action)
	}

	return path, action, nil
}

func transform(content string, edit Edit, existed bool, failOnNoop bool) (string, int, error) {
	action := strings.ToLower(strings.TrimSpace(edit.Action))
	switch action {
	case "replace":
		if edit.Old == "" {
			return "", 0, fmt.Errorf("replace requires old")
		}
		count := strings.Count(content, edit.Old)
		if edit.ReplaceAll {
			if count == 0 && failOnNoop {
				return "", 0, fmt.Errorf("replace_all found no matches")
			}
			return strings.ReplaceAll(content, edit.Old, edit.New), count, nil
		}
		if count != 1 {
			return "", count, fmt.Errorf("replace requires exactly 1 match, got %d", count)
		}
		return strings.Replace(content, edit.Old, edit.New, 1), count, nil
	case "insert_before":
		if edit.Anchor == "" {
			return "", 0, fmt.Errorf("insert_before requires anchor")
		}
		count := strings.Count(content, edit.Anchor)
		if count != 1 {
			return "", count, fmt.Errorf("insert_before requires exactly 1 anchor match, got %d", count)
		}
		return strings.Replace(content, edit.Anchor, edit.New+edit.Anchor, 1), count, nil
	case "insert_after":
		if edit.Anchor == "" {
			return "", 0, fmt.Errorf("insert_after requires anchor")
		}
		count := strings.Count(content, edit.Anchor)
		if count != 1 {
			return "", count, fmt.Errorf("insert_after requires exactly 1 anchor match, got %d", count)
		}
		return strings.Replace(content, edit.Anchor, edit.Anchor+edit.New, 1), count, nil
	case "replace_lines":
		if edit.StartLine <= 0 || edit.EndLine <= 0 || edit.EndLine < edit.StartLine {
			return "", 0, fmt.Errorf("replace_lines requires valid start_line and end_line")
		}
		updated, err := replaceLines(content, edit.StartLine, edit.EndLine, edit.New)
		if err != nil {
			return "", 0, err
		}
		return updated, edit.EndLine - edit.StartLine + 1, nil
	case "append":
		return content + edit.New, 1, nil
	case "write":
		if !existed && !edit.Create {
			return "", 0, fmt.Errorf("write requires create=true when file does not exist")
		}
		return edit.New, 1, nil
	default:
		return "", 0, fmt.Errorf("unsupported action %q", edit.Action)
	}
}

func replaceLines(content string, startLine, endLine int, replacement string) (string, error) {
	lines := splitLines(content)
	if endLine > len(lines) {
		return "", fmt.Errorf("replace_lines range %d-%d exceeds file with %d lines", startLine, endLine, len(lines))
	}

	prefix := strings.Join(lines[:startLine-1], "")
	suffix := strings.Join(lines[endLine:], "")
	return prefix + replacement + suffix, nil
}

func splitLines(content string) []string {
	if content == "" {
		return []string{}
	}

	lines := strings.SplitAfter(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		return lines[:len(lines)-1]
	}
	return lines
}
