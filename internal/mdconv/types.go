package mdconv

import "fmt"

type Result struct {
	Source   string   `json:"source" yaml:"source"`
	Format   string   `json:"format" yaml:"format"`
	Markdown string   `json:"markdown" yaml:"markdown"`
	Warnings []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

type Options struct {
	Name              string
	Format            string
	MaxArchiveEntries int
	ArchiveDepth      int
}

type detector func(name string, data []byte) (string, bool)
type converter func(data []byte, opts convertOptions) (Result, error)

type convertOptions struct {
	Name              string
	Format            string
	MaxArchiveEntries int
	ArchiveDepth      int
}

func (opts Options) normalized() convertOptions {
	maxEntries := opts.MaxArchiveEntries
	if maxEntries <= 0 {
		maxEntries = 64
	}

	depth := opts.ArchiveDepth
	if depth <= 0 {
		depth = 2
	}

	return convertOptions{
		Name:              opts.Name,
		Format:            opts.Format,
		MaxArchiveEntries: maxEntries,
		ArchiveDepth:      depth,
	}
}

func unsupportedFormatError(format string) error {
	if format == "" {
		return fmt.Errorf("could not detect input format")
	}

	return fmt.Errorf("unsupported input format %q", format)
}
