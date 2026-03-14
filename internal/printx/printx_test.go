package printx

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type runnerCall struct {
	Name string
	Args []string
}

type fakeRunner struct {
	calls []runnerCall
	run   func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (runner *fakeRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	runner.calls = append(runner.calls, runnerCall{
		Name: name,
		Args: append([]string(nil), args...),
	})
	if runner.run == nil {
		return nil, nil
	}
	return runner.run(ctx, name, args...)
}

func okLookPath(string) (string, error) {
	return "/usr/bin/tool", nil
}

func TestListParsesConfiguredPrinters(t *testing.T) {
	runner := &fakeRunner{
		run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			switch {
			case name == "lpstat" && reflect.DeepEqual(args, []string{"-p"}):
				return []byte("printer Office is idle. enabled since Fri 13 Mar 2026 09:00:00 AM CET\nprinter Label disabled since Fri 13 Mar 2026 09:05:00 AM CET - paused\n"), nil
			case name == "lpstat" && reflect.DeepEqual(args, []string{"-d"}):
				return []byte("system default destination: Office\n"), nil
			default:
				return nil, fmt.Errorf("unexpected command %s %v", name, args)
			}
		},
	}

	service := NewServiceWithDeps(runner, okLookPath, nil)
	result, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(result.Printers) != 2 {
		t.Fatalf("len(Printers) = %d, want 2", len(result.Printers))
	}
	if !result.Printers[0].Default || result.Printers[0].Name != "Office" {
		t.Fatalf("first printer = %#v, want Office default", result.Printers[0])
	}
	if result.Printers[1].Enabled {
		t.Fatalf("Label Enabled = true, want false")
	}
}

func TestDiscoverRanksSecurePrinters(t *testing.T) {
	runner := &fakeRunner{
		run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if name != "lpinfo" {
				return nil, fmt.Errorf("unexpected command %s", name)
			}
			return []byte(strings.Join([]string{
				"network ipp://printer.local/ipp/print",
				"network dnssd://Office%20Laser._ipps._tcp.local/?uuid=1234",
				"network ipps://secure.local/ipp/print",
				"network socket://legacy-printer",
			}, "\n")), nil
		},
	}

	service := NewServiceWithDeps(runner, okLookPath, nil)
	result, err := service.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if len(result.Printers) != 3 {
		t.Fatalf("len(Printers) = %d, want 3", len(result.Printers))
	}
	if result.Printers[0].Scheme != "ipps" {
		t.Fatalf("first scheme = %q, want ipps", result.Printers[0].Scheme)
	}
	if result.Printers[1].Name != "Office Laser" {
		t.Fatalf("second name = %q, want decoded DNS-SD name", result.Printers[1].Name)
	}
}

func TestEnsureDiscoversAndRepairsQueue(t *testing.T) {
	runner := &fakeRunner{
		run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			switch name {
			case "lpinfo":
				return []byte("network ipps://office.local/ipp/print\nnetwork ipp://office.local/ipp/print\n"), nil
			case "lpadmin", "cupsenable", "cupsaccept":
				return nil, nil
			default:
				return nil, fmt.Errorf("unexpected command %s %v", name, args)
			}
		},
	}

	service := NewServiceWithDeps(runner, okLookPath, nil)
	result, err := service.Ensure(context.Background(), EnsureOptions{
		QueueName:   "office",
		Match:       "office.local",
		MakeDefault: true,
		Description: "Office queue",
		Location:    "Lab",
	})
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	if result.URI != "ipps://office.local/ipp/print" {
		t.Fatalf("URI = %q, want secure ipps URI", result.URI)
	}

	gotCommands := make([]string, 0, len(runner.calls))
	for _, call := range runner.calls {
		gotCommands = append(gotCommands, call.Name+" "+strings.Join(call.Args, " "))
	}
	wantCommands := []string{
		"lpinfo -v",
		"lpadmin -p office -E -v ipps://office.local/ipp/print -m everywhere -D Office queue -L Lab",
		"cupsenable office",
		"cupsaccept office",
		"lpadmin -d office",
	}
	if !reflect.DeepEqual(gotCommands, wantCommands) {
		t.Fatalf("commands = %#v, want %#v", gotCommands, wantCommands)
	}
}

func TestPrintDownloadsURLAndCleansUp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "printer payload")
	}))
	defer server.Close()

	var submittedPath string
	runner := &fakeRunner{
		run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if name != "lp" {
				return nil, fmt.Errorf("unexpected command %s", name)
			}
			submittedPath = args[len(args)-1]
			if _, err := os.Stat(submittedPath); err != nil {
				t.Fatalf("temp file missing during lp run: %v", err)
			}
			return []byte("request id is Office-17 (1 file(s))\n"), nil
		},
	}

	service := NewServiceWithDeps(runner, okLookPath, server.Client())
	result, err := service.Print(context.Background(), PrintOptions{
		Printer:    "Office",
		Source:     server.URL + "/test.txt",
		FitToPage:  true,
		ColorMode:  "monochrome",
		RawOptions: []string{"page-ranges=1-2"},
	})
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !result.Downloaded || result.JobID != 17 {
		t.Fatalf("result = %#v, want downloaded job 17", result)
	}
	if _, err := os.Stat(submittedPath); !os.IsNotExist(err) {
		t.Fatalf("Stat(%s) error = %v, want not exist after cleanup", submittedPath, err)
	}
}

func TestBuildPrintArgsRejectsInvalidCombination(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "doc.pdf")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, _, err := buildPrintArgs(PrintOptions{
		Printer:      "Office",
		Source:       filePath,
		ScalePercent: 90,
		FitToPage:    true,
	}, filePath)
	if err == nil || !strings.Contains(err.Error(), "scale-percent") {
		t.Fatalf("buildPrintArgs() error = %v, want scale-percent validation", err)
	}
}

func TestPrintRejectsUnsupportedURLScheme(t *testing.T) {
	service := NewServiceWithDeps(&fakeRunner{}, okLookPath, nil)

	_, err := service.Print(context.Background(), PrintOptions{
		Printer: "Office",
		Source:  "ftp://printer.local/document.pdf",
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported URL scheme") {
		t.Fatalf("Print() error = %v, want unsupported URL scheme", err)
	}
}
