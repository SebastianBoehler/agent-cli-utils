package smtpx

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultTimeout = 15 * time.Second

func LoadConfig(path string) (Config, error) {
	if strings.TrimSpace(path) == "" {
		return Config{}, nil
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(payload, &config); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	return config, nil
}

func EnvConfig() Config {
	return Config{
		Provider:     strings.TrimSpace(os.Getenv("AGENTSMTP_PROVIDER")),
		Host:         strings.TrimSpace(os.Getenv("AGENTSMTP_HOST")),
		Port:         parseEnvInt("AGENTSMTP_PORT"),
		Security:     strings.TrimSpace(os.Getenv("AGENTSMTP_SECURITY")),
		Auth:         strings.TrimSpace(os.Getenv("AGENTSMTP_AUTH")),
		Username:     strings.TrimSpace(os.Getenv("AGENTSMTP_USERNAME")),
		Password:     os.Getenv("AGENTSMTP_PASSWORD"),
		PasswordEnv:  strings.TrimSpace(os.Getenv("AGENTSMTP_PASSWORD_ENV")),
		PasswordFile: strings.TrimSpace(os.Getenv("AGENTSMTP_PASSWORD_FILE")),
		From:         strings.TrimSpace(os.Getenv("AGENTSMTP_FROM")),
		Timeout:      parseEnvDuration("AGENTSMTP_TIMEOUT"),
	}
}

func Merge(base Config, overrides ...Config) Config {
	merged := base
	for _, override := range overrides {
		if override.Provider != "" {
			merged.Provider = override.Provider
		}
		if override.Host != "" {
			merged.Host = override.Host
		}
		if override.Port != 0 {
			merged.Port = override.Port
		}
		if override.Security != "" {
			merged.Security = override.Security
		}
		if override.Auth != "" {
			merged.Auth = override.Auth
		}
		if override.Username != "" {
			merged.Username = override.Username
		}
		if override.Password != "" {
			merged.Password = override.Password
		}
		if override.PasswordEnv != "" {
			merged.PasswordEnv = override.PasswordEnv
		}
		if override.PasswordFile != "" {
			merged.PasswordFile = override.PasswordFile
		}
		if override.From != "" {
			merged.From = override.From
		}
		if override.Timeout != 0 {
			merged.Timeout = override.Timeout
		}
	}
	return merged
}

func Resolve(config Config) (Config, error) {
	resolved := config
	resolved.Provider = normalizeProvider(resolved.Provider)
	resolved.Security = strings.ToLower(strings.TrimSpace(resolved.Security))
	resolved.Auth = strings.ToLower(strings.TrimSpace(resolved.Auth))
	resolved.Host = strings.TrimSpace(resolved.Host)
	resolved.Username = strings.TrimSpace(resolved.Username)
	resolved.From = strings.TrimSpace(resolved.From)

	if resolved.Provider == "auto" {
		resolved.Provider = inferProvider(resolved)
	}
	if resolved.Provider == "" {
		resolved.Provider = inferProvider(resolved)
	}

	applyProviderDefaults(&resolved)
	applyTransportDefaults(&resolved)

	if resolved.Timeout <= 0 {
		resolved.Timeout = defaultTimeout
	}

	if resolved.From == "" && strings.Contains(resolved.Username, "@") {
		resolved.From = resolved.Username
	}
	if resolved.Username == "" && strings.Contains(resolved.From, "@") {
		resolved.Username = resolved.From
	}

	if resolved.Host == "" {
		return Config{}, fmt.Errorf("smtp host is required")
	}
	if resolved.Port <= 0 {
		return Config{}, fmt.Errorf("smtp port must be greater than zero")
	}
	if !isAllowed(resolved.Security, "starttls", "tls", "plain") {
		return Config{}, fmt.Errorf("unsupported security %q", resolved.Security)
	}
	if !isAllowed(resolved.Auth, "password", "none") {
		return Config{}, fmt.Errorf("unsupported auth %q", resolved.Auth)
	}
	if resolved.From == "" {
		return Config{}, fmt.Errorf("from address is required")
	}
	if _, _, err := normalizeAddress(resolved.From); err != nil {
		return Config{}, fmt.Errorf("invalid from address: %w", err)
	}
	if resolved.Auth != "none" && resolved.Username == "" {
		return Config{}, fmt.Errorf("username is required when auth is enabled")
	}

	return resolved, nil
}

func Inspect(config Config) Profile {
	source, hasSecret := detectSecretSource(config)
	return Profile{
		Operation:    "profile",
		Provider:     config.Provider,
		Host:         config.Host,
		Port:         config.Port,
		Security:     config.Security,
		Auth:         config.Auth,
		Username:     config.Username,
		From:         config.From,
		Timeout:      config.Timeout.String(),
		HasSecret:    hasSecret,
		SecretSource: source,
	}
}

func parseEnvInt(name string) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return 0
	}

	var port int
	_, _ = fmt.Sscanf(value, "%d", &port)
	return port
}

func parseEnvDuration(name string) time.Duration {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return 0
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return duration
}

func detectSecretSource(config Config) (string, bool) {
	switch config.Auth {
	case "password":
		switch {
		case config.Password != "":
			return "inline", true
		case config.PasswordEnv != "":
			_, ok := os.LookupEnv(config.PasswordEnv)
			return "env:" + config.PasswordEnv, ok
		case config.PasswordFile != "":
			payload, err := os.ReadFile(config.PasswordFile)
			if err != nil {
				return "file:" + config.PasswordFile, false
			}
			return "file:" + config.PasswordFile, strings.TrimSpace(string(payload)) != ""
		}
	case "none":
		return "", false
	}
	return "", false
}

func normalizeProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "", "auto", "generic", "gmail", "google-workspace", "google_workspace", "workspace":
		if provider == "google_workspace" || provider == "workspace" {
			return "google-workspace"
		}
		return provider
	default:
		return provider
	}
}

func inferProvider(config Config) string {
	address := strings.ToLower(strings.TrimSpace(config.From))
	if address == "" {
		address = strings.ToLower(strings.TrimSpace(config.Username))
	}
	if strings.HasSuffix(address, "@gmail.com") || strings.HasSuffix(address, "@googlemail.com") {
		return "gmail"
	}
	if strings.EqualFold(strings.TrimSpace(config.Host), "smtp.gmail.com") {
		return "google-workspace"
	}
	return "generic"
}

func applyProviderDefaults(config *Config) {
	switch config.Provider {
	case "gmail", "google-workspace":
		if config.Host == "" {
			config.Host = "smtp.gmail.com"
		}
		if config.Port == 0 {
			config.Port = 587
		}
		if config.Security == "" {
			config.Security = "starttls"
		}
		if config.Auth == "" {
			config.Auth = "password"
		}
	}
}

func applyTransportDefaults(config *Config) {
	if config.Security == "" {
		if config.Port == 465 {
			config.Security = "tls"
		} else {
			config.Security = "starttls"
		}
	}
	if config.Port == 0 {
		if config.Security == "tls" {
			config.Port = 465
		} else {
			config.Port = 587
		}
	}
	if config.Auth == "" {
		config.Auth = defaultAuth(*config)
	}
}

func defaultAuth(config Config) string {
	if hasPassword(config) || config.Username != "" {
		return "password"
	}
	return "none"
}

func hasPassword(config Config) bool {
	return config.Password != "" || config.PasswordEnv != "" || config.PasswordFile != ""
}

func isAllowed(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}
