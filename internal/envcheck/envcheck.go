package envcheck

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Variable struct {
	Name    string `json:"name" yaml:"name"`
	Present bool   `json:"present" yaml:"present"`
	Empty   bool   `json:"empty" yaml:"empty"`
	Length  int    `json:"length,omitempty" yaml:"length,omitempty"`
	Value   string `json:"value,omitempty" yaml:"value,omitempty"`
}

type Report struct {
	AllPresent bool       `json:"all_present" yaml:"all_present"`
	Missing    []string   `json:"missing" yaml:"missing"`
	Empty      []string   `json:"empty" yaml:"empty"`
	Variables  []Variable `json:"variables" yaml:"variables"`
}

func Check(names []string, allowEmpty bool, showValues bool) Report {
	cleanNames := uniqueNames(names)
	report := Report{
		AllPresent: true,
		Missing:    make([]string, 0),
		Empty:      make([]string, 0),
		Variables:  make([]Variable, 0, len(cleanNames)),
	}

	for _, name := range cleanNames {
		value, present := os.LookupEnv(name)
		variable := Variable{
			Name:    name,
			Present: present,
		}

		if !present {
			report.Missing = append(report.Missing, name)
			report.AllPresent = false
			report.Variables = append(report.Variables, variable)
			continue
		}

		variable.Empty = value == ""
		variable.Length = len(value)
		if showValues {
			variable.Value = value
		}

		if variable.Empty && !allowEmpty {
			report.Empty = append(report.Empty, name)
			report.AllPresent = false
		}

		report.Variables = append(report.Variables, variable)
	}

	return report
}

func LoadNames(reader io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	var names []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		names = append(names, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read names: %w", err)
	}

	return names, nil
}

func uniqueNames(names []string) []string {
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}

		set[trimmed] = struct{}{}
	}

	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
