package doctor

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
)

type CheckSpec struct {
	Name        string
	Required    bool
	Description string
	Hint        string
}

type CheckResult struct {
	Name        string `json:"name" yaml:"name"`
	Required    bool   `json:"required" yaml:"required"`
	Present     bool   `json:"present" yaml:"present"`
	Path        string `json:"path,omitempty" yaml:"path,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Hint        string `json:"hint,omitempty" yaml:"hint,omitempty"`
}

type Report struct {
	Profile            string        `json:"profile" yaml:"profile"`
	OS                 string        `json:"os" yaml:"os"`
	Arch               string        `json:"arch" yaml:"arch"`
	AllRequiredPresent bool          `json:"all_required_present" yaml:"all_required_present"`
	Checks             []CheckResult `json:"checks" yaml:"checks"`
}

func AvailableProfiles() []string {
	names := make([]string, 0, len(profileCatalog))
	for name := range profileCatalog {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func Evaluate(profile string, extraCommands []string, goos string, lookPath func(string) (string, error)) (Report, error) {
	if lookPath == nil {
		return Report{}, fmt.Errorf("lookPath resolver is required")
	}

	specs, err := buildChecks(profile, goos)
	if err != nil {
		return Report{}, err
	}

	for _, command := range extraCommands {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		specs = append(specs, CheckSpec{
			Name:        command,
			Required:    true,
			Description: "custom dependency check",
			Hint:        "install the command or add it to PATH",
		})
	}

	report := Report{
		Profile:            profile,
		OS:                 goos,
		Arch:               runtime.GOARCH,
		AllRequiredPresent: true,
		Checks:             make([]CheckResult, 0, len(specs)),
	}

	for _, spec := range specs {
		path, err := lookPath(spec.Name)
		result := CheckResult{
			Name:        spec.Name,
			Required:    spec.Required,
			Present:     err == nil,
			Description: spec.Description,
			Hint:        spec.Hint,
		}
		if err == nil {
			result.Path = path
		}
		if result.Required && !result.Present {
			report.AllRequiredPresent = false
		}
		report.Checks = append(report.Checks, result)
	}

	return report, nil
}

func RenderText(report Report) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "profile: %s\n", report.Profile)
	fmt.Fprintf(&builder, "os: %s\n", report.OS)
	fmt.Fprintf(&builder, "arch: %s\n", report.Arch)
	fmt.Fprintf(&builder, "all_required_present: %t\n", report.AllRequiredPresent)
	for _, check := range report.Checks {
		fmt.Fprintf(
			&builder,
			"%-14s required=%t present=%t",
			check.Name,
			check.Required,
			check.Present,
		)
		if check.Path != "" {
			fmt.Fprintf(&builder, " path=%s", check.Path)
		}
		builder.WriteByte('\n')
		if check.Description != "" {
			fmt.Fprintf(&builder, "  description: %s\n", check.Description)
		}
		if check.Hint != "" && !check.Present {
			fmt.Fprintf(&builder, "  hint: %s\n", check.Hint)
		}
	}
	return builder.String()
}

var profileCatalog = map[string]func(goos string) ([]CheckSpec, error){
	"smb-client": smbClientChecks,
	"ssh-client": sshClientChecks,
	"web-fetch":  webFetchChecks,
}

func buildChecks(profile string, goos string) ([]CheckSpec, error) {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return nil, nil
	}

	builder, ok := profileCatalog[profile]
	if !ok {
		return nil, fmt.Errorf("unknown profile %q", profile)
	}
	return builder(goos)
}

func smbClientChecks(goos string) ([]CheckSpec, error) {
	switch goos {
	case "linux":
		return []CheckSpec{
			{
				Name:        "smbclient",
				Required:    true,
				Description: "SMB/CIFS client for browsing and copying from Windows shares",
				Hint:        "install smbclient or samba-client",
			},
			{
				Name:        "mount.cifs",
				Required:    false,
				Description: "mount helper for SMB/CIFS shares",
				Hint:        "install cifs-utils if the workflow mounts SMB shares",
			},
		}, nil
	case "darwin":
		return []CheckSpec{
			{
				Name:        "smbutil",
				Required:    true,
				Description: "macOS SMB client utility",
				Hint:        "ensure the built-in SMB tools are available on the host",
			},
			{
				Name:        "mount_smbfs",
				Required:    false,
				Description: "mount helper for SMB shares on macOS",
				Hint:        "use mount_smbfs if the workflow mounts shares instead of copying files",
			},
		}, nil
	case "windows":
		return []CheckSpec{
			{
				Name:        "powershell.exe",
				Required:    true,
				Description: "Windows shell often used for SMB automation and diagnostics",
				Hint:        "ensure PowerShell is available in PATH",
			},
			{
				Name:        "net.exe",
				Required:    true,
				Description: "Windows network utility for share mapping and SMB checks",
				Hint:        "ensure net.exe is available in PATH",
			},
		}, nil
	default:
		return nil, fmt.Errorf("smb-client profile does not support %s", goos)
	}
}

func sshClientChecks(goos string) ([]CheckSpec, error) {
	switch goos {
	case "linux", "darwin":
		return []CheckSpec{
			{
				Name:        "ssh",
				Required:    true,
				Description: "SSH client",
				Hint:        "install openssh-client",
			},
			{
				Name:        "scp",
				Required:    false,
				Description: "SSH file copy utility",
				Hint:        "install openssh-client if file copy is needed",
			},
		}, nil
	case "windows":
		return []CheckSpec{
			{
				Name:        "ssh.exe",
				Required:    true,
				Description: "OpenSSH client on Windows",
				Hint:        "enable the OpenSSH Client optional feature",
			},
			{
				Name:        "scp.exe",
				Required:    false,
				Description: "OpenSSH secure copy utility",
				Hint:        "enable the OpenSSH Client optional feature if copy is needed",
			},
		}, nil
	default:
		return nil, fmt.Errorf("ssh-client profile does not support %s", goos)
	}
}

func webFetchChecks(goos string) ([]CheckSpec, error) {
	switch goos {
	case "linux", "darwin":
		return []CheckSpec{
			{
				Name:        "curl",
				Required:    true,
				Description: "HTTP fetch client",
				Hint:        "install curl",
			},
		}, nil
	case "windows":
		return []CheckSpec{
			{
				Name:        "curl.exe",
				Required:    false,
				Description: "HTTP fetch client on Windows",
				Hint:        "install curl or use PowerShell Invoke-WebRequest",
			},
			{
				Name:        "powershell.exe",
				Required:    true,
				Description: "PowerShell for HTTP requests and diagnostics",
				Hint:        "ensure PowerShell is available in PATH",
			},
		}, nil
	default:
		return nil, fmt.Errorf("web-fetch profile does not support %s", goos)
	}
}
