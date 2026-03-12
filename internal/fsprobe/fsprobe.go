package fsprobe

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Options struct {
	MaxDepth      int
	IncludeHidden bool
	Hash          bool
	PreviewLines  int
	MaxEntries    int
}

type Entry struct {
	Path    string   `json:"path" yaml:"path"`
	RelPath string   `json:"rel_path" yaml:"rel_path"`
	Name    string   `json:"name" yaml:"name"`
	Type    string   `json:"type" yaml:"type"`
	Size    int64    `json:"size" yaml:"size"`
	Mode    string   `json:"mode" yaml:"mode"`
	ModTime string   `json:"mod_time" yaml:"mod_time"`
	SHA256  string   `json:"sha256,omitempty" yaml:"sha256,omitempty"`
	Preview []string `json:"preview,omitempty" yaml:"preview,omitempty"`
	Error   string   `json:"error,omitempty" yaml:"error,omitempty"`
}

type Result struct {
	Root        string  `json:"root" yaml:"root"`
	Entries     []Entry `json:"entries" yaml:"entries"`
	Directories int     `json:"directories" yaml:"directories"`
	Files       int     `json:"files" yaml:"files"`
	Symlinks    int     `json:"symlinks" yaml:"symlinks"`
	Truncated   bool    `json:"truncated" yaml:"truncated"`
}

func Probe(root string, options Options) (Result, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Result{}, fmt.Errorf("resolve root: %w", err)
	}

	info, err := os.Lstat(absoluteRoot)
	if err != nil {
		return Result{}, fmt.Errorf("stat root: %w", err)
	}

	if options.MaxDepth < 0 {
		options.MaxDepth = 0
	}

	if options.MaxEntries <= 0 {
		options.MaxEntries = 1000
	}

	result := Result{
		Root:    absoluteRoot,
		Entries: make([]Entry, 0, min(options.MaxEntries, 128)),
	}

	appendEntry := func(path string, entryInfo fs.FileInfo) {
		if len(result.Entries) >= options.MaxEntries {
			result.Truncated = true
			return
		}

		relativePath, relErr := filepath.Rel(absoluteRoot, path)
		if relErr != nil || relativePath == "." {
			relativePath = "."
		}

		entry := Entry{
			Path:    path,
			RelPath: relativePath,
			Name:    entryInfo.Name(),
			Type:    fileType(entryInfo.Mode()),
			Size:    entryInfo.Size(),
			Mode:    entryInfo.Mode().String(),
			ModTime: entryInfo.ModTime().UTC().Format(time.RFC3339),
		}

		switch entry.Type {
		case "dir":
			result.Directories++
		case "file":
			result.Files++
		case "symlink":
			result.Symlinks++
		}

		if entry.Type == "file" {
			if options.Hash {
				sum, hashErr := fileSHA256(path)
				if hashErr != nil {
					entry.Error = hashErr.Error()
				} else {
					entry.SHA256 = sum
				}
			}

			if options.PreviewLines > 0 {
				preview, previewErr := previewFile(path, options.PreviewLines)
				if previewErr != nil && entry.Error == "" {
					entry.Error = previewErr.Error()
				} else {
					entry.Preview = preview
				}
			}
		}

		result.Entries = append(result.Entries, entry)
	}

	if !info.IsDir() {
		appendEntry(absoluteRoot, info)
		return result, nil
	}

	walkErr := filepath.WalkDir(absoluteRoot, func(path string, dirEntry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if result.Truncated {
			if dirEntry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relativePath, relErr := filepath.Rel(absoluteRoot, path)
		if relErr != nil {
			return relErr
		}

		if relativePath != "." {
			if !options.IncludeHidden && hasHiddenSegment(relativePath) {
				if dirEntry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if depth(relativePath) > options.MaxDepth {
				if dirEntry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		entryInfo, err := dirEntry.Info()
		if err != nil {
			return err
		}

		appendEntry(path, entryInfo)
		return nil
	})

	if walkErr != nil {
		return Result{}, walkErr
	}

	return result, nil
}

func RenderText(result Result) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "root: %s\n", result.Root)
	fmt.Fprintf(&builder, "entries: %d files=%d dirs=%d symlinks=%d truncated=%t\n", len(result.Entries), result.Files, result.Directories, result.Symlinks, result.Truncated)

	for _, entry := range result.Entries {
		fmt.Fprintf(&builder, "%-7s %10d %s\n", entry.Type, entry.Size, entry.RelPath)
		if len(entry.Preview) > 0 {
			for _, line := range entry.Preview {
				fmt.Fprintf(&builder, "  > %s\n", line)
			}
		}
		if entry.Error != "" {
			fmt.Fprintf(&builder, "  ! %s\n", entry.Error)
		}
	}

	return builder.String()
}

func fileType(mode fs.FileMode) string {
	switch {
	case mode.IsDir():
		return "dir"
	case mode&fs.ModeSymlink != 0:
		return "symlink"
	case mode.IsRegular():
		return "file"
	default:
		return "other"
	}
}

func hasHiddenSegment(path string) bool {
	for _, part := range strings.Split(path, string(os.PathSeparator)) {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}

func depth(path string) int {
	if path == "." || path == "" {
		return 0
	}
	return strings.Count(path, string(os.PathSeparator)) + 1
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func previewFile(path string, maxLines int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("preview %s: %w", path, err)
	}
	defer file.Close()

	head := make([]byte, 512)
	count, err := file.Read(head)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("preview %s: %w", path, err)
	}

	if bytesLookBinary(head[:count]) {
		return nil, nil
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("preview %s: %w", path, err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	lines := make([]string, 0, maxLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) >= maxLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("preview %s: %w", path, err)
	}

	return lines, nil
}

func bytesLookBinary(payload []byte) bool {
	for _, item := range payload {
		if item == 0 {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
