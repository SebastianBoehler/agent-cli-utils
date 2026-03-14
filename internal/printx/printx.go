package printx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type Service struct {
	runner     CommandRunner
	lookPath   func(string) (string, error)
	httpClient *http.Client
}

type ListResult struct {
	Printers []Printer `json:"printers" yaml:"printers"`
}

type Printer struct {
	Name    string `json:"name" yaml:"name"`
	Default bool   `json:"default" yaml:"default"`
	Enabled bool   `json:"enabled" yaml:"enabled"`
	State   string `json:"state" yaml:"state"`
}

type DiscoverResult struct {
	Printers []DiscoveredPrinter `json:"printers" yaml:"printers"`
}

type DiscoveredPrinter struct {
	Name   string `json:"name" yaml:"name"`
	URI    string `json:"uri" yaml:"uri"`
	Scheme string `json:"scheme" yaml:"scheme"`
	Secure bool   `json:"secure" yaml:"secure"`
}

type EnsureOptions struct {
	QueueName   string
	URI         string
	Match       string
	MakeDefault bool
	Description string
	Location    string
}

type EnsureResult struct {
	Queue       string             `json:"queue" yaml:"queue"`
	URI         string             `json:"uri" yaml:"uri"`
	Default     bool               `json:"default" yaml:"default"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Location    string             `json:"location,omitempty" yaml:"location,omitempty"`
	Discovered  *DiscoveredPrinter `json:"discovered,omitempty" yaml:"discovered,omitempty"`
}

type PrintOptions struct {
	Printer      string
	Source       string
	Copies       int
	Duplex       bool
	Media        string
	Orientation  string
	Position     string
	ScalePercent int
	FitToPage    bool
	FillPage     bool
	ColorMode    string
	RawOptions   []string
	JobName      string
}

type PrintResult struct {
	Printer    string   `json:"printer" yaml:"printer"`
	Source     string   `json:"source" yaml:"source"`
	SourceType string   `json:"source_type" yaml:"source_type"`
	FilePath   string   `json:"file_path" yaml:"file_path"`
	Downloaded bool     `json:"downloaded" yaml:"downloaded"`
	RequestID  string   `json:"request_id,omitempty" yaml:"request_id,omitempty"`
	JobID      int      `json:"job_id,omitempty" yaml:"job_id,omitempty"`
	Options    []string `json:"options,omitempty" yaml:"options,omitempty"`
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

func NewService() *Service {
	return &Service{
		runner:     execRunner{},
		lookPath:   exec.LookPath,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func NewServiceWithDeps(runner CommandRunner, lookPath func(string) (string, error), client *http.Client) *Service {
	if runner == nil {
		runner = execRunner{}
	}
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	return &Service{
		runner:     runner,
		lookPath:   lookPath,
		httpClient: client,
	}
}

func (s *Service) List(ctx context.Context) (ListResult, error) {
	if err := s.requireTools("lpstat"); err != nil {
		return ListResult{}, err
	}

	printersOut, err := s.call(ctx, "lpstat", "-p")
	if err != nil {
		return ListResult{}, commandError("lpstat", []string{"-p"}, printersOut, err)
	}

	defaultOut, defaultErr := s.call(ctx, "lpstat", "-d")
	defaultName := ""
	if defaultErr == nil {
		defaultName = parseDefaultPrinter(defaultOut)
	} else if !isNoDefaultPrinterOutput(defaultOut) {
		return ListResult{}, commandError("lpstat", []string{"-d"}, defaultOut, defaultErr)
	}

	printers := parseLPStatPrinters(printersOut)
	for index := range printers {
		printers[index].Default = printers[index].Name == defaultName
	}

	sort.Slice(printers, func(i, j int) bool {
		if printers[i].Default != printers[j].Default {
			return printers[i].Default
		}
		return printers[i].Name < printers[j].Name
	})

	return ListResult{Printers: printers}, nil
}

func (s *Service) Discover(ctx context.Context) (DiscoverResult, error) {
	if err := s.requireTools("lpinfo"); err != nil {
		return DiscoverResult{}, err
	}

	out, err := s.call(ctx, "lpinfo", "-v")
	if err != nil {
		return DiscoverResult{}, commandError("lpinfo", []string{"-v"}, out, err)
	}

	printers := parseLPInfo(out)
	sort.Slice(printers, func(i, j int) bool {
		leftRank := discoveryRank(printers[i])
		rightRank := discoveryRank(printers[j])
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if printers[i].Name != printers[j].Name {
			return printers[i].Name < printers[j].Name
		}
		return printers[i].URI < printers[j].URI
	})

	return DiscoverResult{Printers: printers}, nil
}

func (s *Service) Ensure(ctx context.Context, options EnsureOptions) (EnsureResult, error) {
	queue := strings.TrimSpace(options.QueueName)
	if queue == "" {
		return EnsureResult{}, fmt.Errorf("printer queue name is required")
	}

	var discovered *DiscoveredPrinter
	uri := strings.TrimSpace(options.URI)
	if uri == "" {
		match := strings.TrimSpace(options.Match)
		if match == "" {
			return EnsureResult{}, fmt.Errorf("ensure requires -uri or -match")
		}

		discovery, err := s.Discover(ctx)
		if err != nil {
			return EnsureResult{}, err
		}

		selected := selectDiscoveredPrinter(discovery.Printers, match)
		if selected == nil {
			return EnsureResult{}, fmt.Errorf("no discovered printer matched %q", match)
		}
		discovered = selected
		uri = selected.URI
	}

	if err := validateQueueURI(uri); err != nil {
		return EnsureResult{}, err
	}
	if err := s.requireTools("lpadmin", "cupsenable", "cupsaccept"); err != nil {
		return EnsureResult{}, err
	}

	lpadminArgs := []string{"-p", queue, "-E", "-v", uri, "-m", "everywhere"}
	if description := strings.TrimSpace(options.Description); description != "" {
		lpadminArgs = append(lpadminArgs, "-D", description)
	}
	if location := strings.TrimSpace(options.Location); location != "" {
		lpadminArgs = append(lpadminArgs, "-L", location)
	}

	if out, err := s.call(ctx, "lpadmin", lpadminArgs...); err != nil {
		return EnsureResult{}, commandError("lpadmin", lpadminArgs, out, err)
	}
	if out, err := s.call(ctx, "cupsenable", queue); err != nil {
		return EnsureResult{}, commandError("cupsenable", []string{queue}, out, err)
	}
	if out, err := s.call(ctx, "cupsaccept", queue); err != nil {
		return EnsureResult{}, commandError("cupsaccept", []string{queue}, out, err)
	}
	if options.MakeDefault {
		if out, err := s.call(ctx, "lpadmin", "-d", queue); err != nil {
			return EnsureResult{}, commandError("lpadmin", []string{"-d", queue}, out, err)
		}
	}

	return EnsureResult{
		Queue:       queue,
		URI:         uri,
		Default:     options.MakeDefault,
		Description: strings.TrimSpace(options.Description),
		Location:    strings.TrimSpace(options.Location),
		Discovered:  discovered,
	}, nil
}

func (s *Service) Print(ctx context.Context, options PrintOptions) (result PrintResult, err error) {
	if err := validatePrintOptions(options); err != nil {
		return PrintResult{}, err
	}
	if err := s.requireTools("lp"); err != nil {
		return PrintResult{}, err
	}

	filePath, sourceType, cleanup, err := s.prepareSource(ctx, options.Source)
	if err != nil {
		return PrintResult{}, err
	}
	if cleanup != nil {
		defer func() {
			cleanupErr := cleanup()
			if err == nil && cleanupErr != nil {
				err = cleanupErr
			}
		}()
	}

	args, normalizedOptions, err := buildPrintArgs(options, filePath)
	if err != nil {
		return PrintResult{}, err
	}

	out, runErr := s.call(ctx, "lp", args...)
	if runErr != nil {
		return PrintResult{}, commandError("lp", args, out, runErr)
	}

	requestID, jobID := parseRequestID(out)
	return PrintResult{
		Printer:    strings.TrimSpace(options.Printer),
		Source:     strings.TrimSpace(options.Source),
		SourceType: sourceType,
		FilePath:   filePath,
		Downloaded: sourceType == "url",
		RequestID:  requestID,
		JobID:      jobID,
		Options:    normalizedOptions,
	}, nil
}

func RenderListText(result ListResult) string {
	var builder strings.Builder
	for _, printer := range result.Printers {
		fmt.Fprintf(
			&builder,
			"%s default=%t enabled=%t state=%s\n",
			printer.Name,
			printer.Default,
			printer.Enabled,
			printer.State,
		)
	}
	return builder.String()
}

func RenderListTSV(result ListResult) string {
	var builder strings.Builder
	builder.WriteString("name\tdefault\tenabled\tstate\n")
	for _, printer := range result.Printers {
		fmt.Fprintf(
			&builder,
			"%s\t%t\t%t\t%s\n",
			printer.Name,
			printer.Default,
			printer.Enabled,
			printer.State,
		)
	}
	return builder.String()
}

func RenderDiscoverText(result DiscoverResult) string {
	var builder strings.Builder
	for _, printer := range result.Printers {
		fmt.Fprintf(
			&builder,
			"%s scheme=%s secure=%t uri=%s\n",
			printer.Name,
			printer.Scheme,
			printer.Secure,
			printer.URI,
		)
	}
	return builder.String()
}

func RenderDiscoverTSV(result DiscoverResult) string {
	var builder strings.Builder
	builder.WriteString("name\tscheme\tsecure\turi\n")
	for _, printer := range result.Printers {
		fmt.Fprintf(
			&builder,
			"%s\t%s\t%t\t%s\n",
			printer.Name,
			printer.Scheme,
			printer.Secure,
			printer.URI,
		)
	}
	return builder.String()
}

func RenderEnsureText(result EnsureResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "queue: %s\n", result.Queue)
	fmt.Fprintf(&builder, "uri: %s\n", result.URI)
	fmt.Fprintf(&builder, "default: %t\n", result.Default)
	if result.Description != "" {
		fmt.Fprintf(&builder, "description: %s\n", result.Description)
	}
	if result.Location != "" {
		fmt.Fprintf(&builder, "location: %s\n", result.Location)
	}
	if result.Discovered != nil {
		fmt.Fprintf(&builder, "discovered_name: %s\n", result.Discovered.Name)
		fmt.Fprintf(&builder, "discovered_scheme: %s\n", result.Discovered.Scheme)
	}
	return builder.String()
}

func RenderEnsureTSV(result EnsureResult) string {
	var builder strings.Builder
	builder.WriteString("queue\turi\tdefault\tdiscovered_name\tdiscovered_scheme\n")
	discoveredName := ""
	discoveredScheme := ""
	if result.Discovered != nil {
		discoveredName = result.Discovered.Name
		discoveredScheme = result.Discovered.Scheme
	}
	fmt.Fprintf(
		&builder,
		"%s\t%s\t%t\t%s\t%s\n",
		result.Queue,
		result.URI,
		result.Default,
		discoveredName,
		discoveredScheme,
	)
	return builder.String()
}

func RenderPrintText(result PrintResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "printer: %s\n", result.Printer)
	fmt.Fprintf(&builder, "source: %s\n", result.Source)
	fmt.Fprintf(&builder, "source_type: %s\n", result.SourceType)
	fmt.Fprintf(&builder, "file_path: %s\n", result.FilePath)
	fmt.Fprintf(&builder, "downloaded: %t\n", result.Downloaded)
	if result.RequestID != "" {
		fmt.Fprintf(&builder, "request_id: %s\n", result.RequestID)
	}
	if result.JobID != 0 {
		fmt.Fprintf(&builder, "job_id: %d\n", result.JobID)
	}
	if len(result.Options) > 0 {
		fmt.Fprintf(&builder, "options: %s\n", strings.Join(result.Options, ","))
	}
	return builder.String()
}

func RenderPrintTSV(result PrintResult) string {
	var builder strings.Builder
	builder.WriteString("printer\tsource\tsource_type\tfile_path\tdownloaded\trequest_id\tjob_id\toptions\n")
	fmt.Fprintf(
		&builder,
		"%s\t%s\t%s\t%s\t%t\t%s\t%d\t%s\n",
		result.Printer,
		result.Source,
		result.SourceType,
		result.FilePath,
		result.Downloaded,
		result.RequestID,
		result.JobID,
		strings.Join(result.Options, ","),
	)
	return builder.String()
}

func parseLPStatPrinters(output string) []Printer {
	lines := strings.Split(output, "\n")
	printers := make([]Printer, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "printer ") {
			continue
		}

		rest := strings.TrimSpace(strings.TrimPrefix(line, "printer "))
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}

		name := fields[0]
		state := strings.TrimSpace(strings.TrimPrefix(rest, name))
		printers = append(printers, Printer{
			Name:    name,
			Enabled: !strings.HasPrefix(strings.ToLower(state), "disabled"),
			State:   state,
		})
	}
	return printers
}

func parseDefaultPrinter(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "system default destination:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "system default destination:"))
		}
	}
	return ""
}

func isNoDefaultPrinterOutput(output string) bool {
	text := strings.ToLower(strings.TrimSpace(output))
	return strings.Contains(text, "no system default destination")
}

func parseLPInfo(output string) []DiscoveredPrinter {
	lines := strings.Split(output, "\n")
	printers := make([]DiscoveredPrinter, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		uri := fields[1]
		scheme := discoveryScheme(uri)
		if !supportedDiscoveryScheme(scheme) {
			continue
		}
		if _, exists := seen[uri]; exists {
			continue
		}
		seen[uri] = struct{}{}

		printers = append(printers, DiscoveredPrinter{
			Name:   discoveryName(uri),
			URI:    uri,
			Scheme: scheme,
			Secure: isSecureDiscovery(scheme, uri),
		})
	}
	return printers
}

func supportedDiscoveryScheme(scheme string) bool {
	switch scheme {
	case "ipp", "ipps", "dnssd":
		return true
	default:
		return false
	}
}

func discoveryScheme(rawURI string) string {
	if strings.HasPrefix(rawURI, "dnssd://") {
		return "dnssd"
	}

	parsed, err := neturl.Parse(rawURI)
	if err == nil && parsed.Scheme != "" {
		return strings.ToLower(parsed.Scheme)
	}
	return ""
}

func isSecureDiscovery(scheme string, rawURI string) bool {
	if scheme == "ipps" {
		return true
	}
	if scheme == "dnssd" {
		return strings.Contains(strings.ToLower(rawURI), "._ipps.")
	}
	return false
}

func discoveryName(rawURI string) string {
	if strings.HasPrefix(rawURI, "dnssd://") {
		host := strings.TrimPrefix(rawURI, "dnssd://")
		host = strings.SplitN(host, "/", 2)[0]
		host = strings.TrimSuffix(host, ".local")
		host = strings.SplitN(host, "._", 2)[0]
		if decoded, err := neturl.PathUnescape(host); err == nil && strings.TrimSpace(decoded) != "" {
			return decoded
		}
		return host
	}

	parsed, err := neturl.Parse(rawURI)
	if err != nil {
		return rawURI
	}

	switch strings.ToLower(parsed.Scheme) {
	case "ipp", "ipps":
		if path := strings.Trim(parsed.Path, "/"); path != "" {
			if segment := lastPathSegment(path); segment != "" && segment != "print" {
				if decoded, err := neturl.PathUnescape(segment); err == nil && decoded != "" {
					return decoded
				}
				return segment
			}
		}
		if host := parsed.Hostname(); host != "" {
			return host
		}
	}

	if parsed.Hostname() != "" {
		return parsed.Hostname()
	}
	return rawURI
}

func lastPathSegment(path string) string {
	parts := strings.Split(path, "/")
	for index := len(parts) - 1; index >= 0; index-- {
		part := strings.TrimSpace(parts[index])
		if part != "" {
			return part
		}
	}
	return ""
}

func discoveryRank(printer DiscoveredPrinter) int {
	switch printer.Scheme {
	case "ipps":
		return 0
	case "dnssd":
		if printer.Secure {
			return 1
		}
		return 3
	case "ipp":
		return 2
	default:
		return 4
	}
}

func selectDiscoveredPrinter(printers []DiscoveredPrinter, match string) *DiscoveredPrinter {
	match = strings.TrimSpace(match)
	if match == "" {
		return nil
	}

	lowerMatch := strings.ToLower(match)
	for index := range printers {
		if printers[index].Name == match {
			return &printers[index]
		}
	}
	for index := range printers {
		if strings.EqualFold(printers[index].Name, match) {
			return &printers[index]
		}
	}
	for index := range printers {
		if printers[index].URI == match {
			return &printers[index]
		}
	}
	for index := range printers {
		if strings.Contains(strings.ToLower(printers[index].Name), lowerMatch) {
			return &printers[index]
		}
	}
	for index := range printers {
		if strings.Contains(strings.ToLower(printers[index].URI), lowerMatch) {
			return &printers[index]
		}
	}
	return nil
}

func validateQueueURI(rawURI string) error {
	parsed, err := neturl.Parse(strings.TrimSpace(rawURI))
	if err != nil {
		return fmt.Errorf("invalid printer URI %q: %w", rawURI, err)
	}
	if parsed.Scheme == "" {
		return fmt.Errorf("printer URI must include a scheme")
	}
	return nil
}

func validatePrintOptions(options PrintOptions) error {
	if strings.TrimSpace(options.Printer) == "" {
		return fmt.Errorf("printer is required")
	}
	if strings.TrimSpace(options.Source) == "" {
		return fmt.Errorf("input file or URL is required")
	}

	copies := options.Copies
	if copies == 0 {
		copies = 1
	}
	if copies < 1 {
		return fmt.Errorf("copies must be at least 1")
	}
	if options.ScalePercent < 0 {
		return fmt.Errorf("scale-percent must be positive")
	}
	if options.ScalePercent > 0 && options.FitToPage {
		return fmt.Errorf("scale-percent cannot be combined with fit-to-page")
	}
	if options.ScalePercent > 0 && options.FillPage {
		return fmt.Errorf("scale-percent cannot be combined with fill-page")
	}
	if options.FitToPage && options.FillPage {
		return fmt.Errorf("fit-to-page cannot be combined with fill-page")
	}

	switch orientation := strings.TrimSpace(strings.ToLower(options.Orientation)); orientation {
	case "", "portrait", "landscape", "reverse-landscape", "reverse-portrait":
	default:
		return fmt.Errorf("unsupported orientation %q", options.Orientation)
	}

	switch colorMode := strings.TrimSpace(strings.ToLower(options.ColorMode)); colorMode {
	case "", "auto", "color", "monochrome", "bi-level":
	default:
		return fmt.Errorf("unsupported color-mode %q", options.ColorMode)
	}

	for _, item := range options.RawOptions {
		if !strings.Contains(item, "=") {
			return fmt.Errorf("raw option %q must use name=value", item)
		}
	}

	return nil
}

func buildPrintArgs(options PrintOptions, filePath string) ([]string, []string, error) {
	if err := validatePrintOptions(options); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(filePath) == "" {
		return nil, nil, fmt.Errorf("file path is required")
	}

	copies := options.Copies
	if copies == 0 {
		copies = 1
	}

	args := []string{"-d", strings.TrimSpace(options.Printer)}
	normalizedOptions := make([]string, 0, len(options.RawOptions)+8)
	if copies > 1 {
		args = append(args, "-n", strconv.Itoa(copies))
	}
	if strings.TrimSpace(options.JobName) != "" {
		args = append(args, "-t", strings.TrimSpace(options.JobName))
	}
	if options.Duplex {
		normalizedOptions = append(normalizedOptions, "sides=two-sided-long-edge")
	}
	if media := strings.TrimSpace(options.Media); media != "" {
		normalizedOptions = append(normalizedOptions, "media="+media)
	}
	if orientation := strings.TrimSpace(strings.ToLower(options.Orientation)); orientation != "" {
		normalizedOptions = append(normalizedOptions, "orientation-requested="+orientationValue(orientation))
	}
	if position := strings.TrimSpace(options.Position); position != "" {
		normalizedOptions = append(normalizedOptions, "position="+position)
	}
	if options.ScalePercent > 0 {
		normalizedOptions = append(normalizedOptions, "scaling="+strconv.Itoa(options.ScalePercent))
	}
	if options.FitToPage {
		normalizedOptions = append(normalizedOptions, "fit-to-page")
	}
	if options.FillPage {
		normalizedOptions = append(normalizedOptions, "print-scaling=fill")
	}
	if colorMode := strings.TrimSpace(strings.ToLower(options.ColorMode)); colorMode != "" && colorMode != "auto" {
		normalizedOptions = append(normalizedOptions, "print-color-mode="+colorMode)
	}
	normalizedOptions = append(normalizedOptions, options.RawOptions...)

	for _, option := range normalizedOptions {
		args = append(args, "-o", option)
	}
	args = append(args, filePath)
	return args, normalizedOptions, nil
}

func orientationValue(orientation string) string {
	switch orientation {
	case "portrait":
		return "3"
	case "landscape":
		return "4"
	case "reverse-landscape":
		return "5"
	case "reverse-portrait":
		return "6"
	default:
		return orientation
	}
}

func parseRequestID(output string) (string, int) {
	matches := requestIDPattern.FindStringSubmatch(output)
	if len(matches) != 2 {
		return "", 0
	}

	requestID := matches[1]
	jobID := 0
	if jobMatches := requestIDSuffixPattern.FindStringSubmatch(requestID); len(jobMatches) == 2 {
		if parsed, err := strconv.Atoi(jobMatches[1]); err == nil {
			jobID = parsed
		}
	}
	return requestID, jobID
}

func (s *Service) prepareSource(ctx context.Context, source string) (string, string, func() error, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", "", nil, fmt.Errorf("input file or URL is required")
	}

	if parsed, err := neturl.Parse(source); err == nil && parsed.Scheme != "" {
		scheme := strings.ToLower(parsed.Scheme)
		switch scheme {
		case "http", "https":
			if parsed.Host == "" {
				return "", "", nil, fmt.Errorf("invalid URL %q", source)
			}
		default:
			if strings.Contains(source, "://") {
				return "", "", nil, fmt.Errorf("unsupported URL scheme %q", parsed.Scheme)
			}
		}
	}

	if isHTTPURL(source) {
		path, cleanup, err := s.download(ctx, source)
		if err != nil {
			return "", "", nil, err
		}
		return path, "url", cleanup, nil
	}

	path := filepath.Clean(source)
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", nil, fmt.Errorf("input file %s does not exist", path)
		}
		return "", "", nil, fmt.Errorf("stat %s: %w", path, err)
	}
	if info.IsDir() {
		return "", "", nil, fmt.Errorf("input path %s is a directory", path)
	}

	return path, "file", nil, nil
}

func (s *Service) download(ctx context.Context, rawURL string) (string, func() error, error) {
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return "", nil, fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if !isHTTPURL(rawURL) {
		return "", nil, fmt.Errorf("unsupported URL scheme %q", parsed.Scheme)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("build download request: %w", err)
	}

	response, err := s.httpClient.Do(request)
	if err != nil {
		return "", nil, fmt.Errorf("download %s: %w", rawURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", nil, fmt.Errorf("download %s: unexpected HTTP status %d", rawURL, response.StatusCode)
	}

	pattern := "agentprint-*"
	if extension := safeExtension(parsed.Path); extension != "" {
		pattern += extension
	}

	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", nil, fmt.Errorf("create temp file: %w", err)
	}
	path := file.Name()

	if _, err := io.Copy(file, response.Body); err != nil {
		file.Close()
		os.Remove(path)
		return "", nil, fmt.Errorf("write temp file: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", nil, fmt.Errorf("close temp file: %w", err)
	}

	return path, func() error {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove temp file %s: %w", path, err)
		}
		return nil
	}, nil
}

func safeExtension(rawPath string) string {
	extension := filepath.Ext(rawPath)
	if len(extension) == 0 || len(extension) > 16 || strings.Contains(extension, string(filepath.Separator)) {
		return ""
	}
	return extension
}

func isHTTPURL(value string) bool {
	parsed, err := neturl.Parse(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	return (scheme == "http" || scheme == "https") && parsed.Host != ""
}

func (s *Service) requireTools(names ...string) error {
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			continue
		}
		if _, err := s.lookPath(name); err != nil {
			return fmt.Errorf("required host tool missing: %s", name)
		}
	}
	return nil
}

func (s *Service) call(ctx context.Context, name string, args ...string) (string, error) {
	output, err := s.runner.Run(ctx, name, args...)
	return string(output), err
}

func commandError(name string, args []string, output string, err error) error {
	command := name
	if len(args) > 0 {
		command += " " + strings.Join(args, " ")
	}
	if trimmed := strings.TrimSpace(output); trimmed != "" {
		return fmt.Errorf("%s: %s", command, trimmed)
	}
	return fmt.Errorf("%s: %w", command, err)
}

var (
	requestIDPattern       = regexp.MustCompile(`request id is ([^ ]+)`)
	requestIDSuffixPattern = regexp.MustCompile(`-([0-9]+)$`)
)
