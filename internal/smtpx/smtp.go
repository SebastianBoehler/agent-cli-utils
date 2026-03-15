package smtpx

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
	"strings"
)

func Check(config Config) (Result, error) {
	client, err := dial(config)
	if err != nil {
		return Result{}, err
	}
	defer client.Close()

	if err := authenticate(client, config); err != nil {
		return Result{}, err
	}
	if err := client.Noop(); err != nil {
		return Result{}, fmt.Errorf("smtp noop failed: %w", err)
	}
	if err := client.Quit(); err != nil {
		return Result{}, fmt.Errorf("smtp quit failed: %w", err)
	}

	return Result{
		Operation: "test",
		Status:    "ok",
		Provider:  config.Provider,
		Host:      config.Host,
		Port:      config.Port,
		Security:  config.Security,
		Auth:      config.Auth,
		Username:  config.Username,
		From:      config.From,
	}, nil
}

func Send(config Config, message Message) (Result, error) {
	_, envelopeFrom, err := normalizeAddress(config.From)
	if err != nil {
		return Result{}, fmt.Errorf("invalid from address: %w", err)
	}

	payload, recipients, err := BuildMessage(config.From, message)
	if err != nil {
		return Result{}, err
	}

	client, err := dial(config)
	if err != nil {
		return Result{}, err
	}
	defer client.Close()

	if err := authenticate(client, config); err != nil {
		return Result{}, err
	}
	if err := client.Mail(envelopeFrom); err != nil {
		return Result{}, fmt.Errorf("smtp MAIL FROM failed: %w", err)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return Result{}, fmt.Errorf("smtp RCPT TO failed for %s: %w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return Result{}, fmt.Errorf("smtp DATA failed: %w", err)
	}
	if _, err := writer.Write(payload); err != nil {
		return Result{}, fmt.Errorf("write smtp message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return Result{}, fmt.Errorf("close smtp message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return Result{}, fmt.Errorf("smtp quit failed: %w", err)
	}

	return Result{
		Operation:    "send",
		Status:       "sent",
		Provider:     config.Provider,
		Host:         config.Host,
		Port:         config.Port,
		Security:     config.Security,
		Auth:         config.Auth,
		Username:     config.Username,
		From:         config.From,
		To:           compactStrings(message.To),
		Cc:           compactStrings(message.Cc),
		Bcc:          compactStrings(message.Bcc),
		Subject:      strings.TrimSpace(message.Subject),
		MessageBytes: len(payload),
	}, nil
}

func ReadBody(inline string, filePath string) (string, error) {
	if strings.TrimSpace(inline) != "" {
		return inline, nil
	}
	if strings.TrimSpace(filePath) == "" {
		if stdinHasData() {
			payload, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("read stdin: %w", err)
			}
			return string(payload), nil
		}
		return "", nil
	}
	if filePath == "-" {
		payload, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(payload), nil
	}

	payload, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read body %s: %w", filePath, err)
	}
	return string(payload), nil
}

func dial(config Config) (*smtp.Client, error) {
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	dialer := net.Dialer{Timeout: config.Timeout}

	switch config.Security {
	case "tls":
		connection, err := tls.DialWithDialer(&dialer, "tcp", address, &tls.Config{ServerName: config.Host})
		if err != nil {
			return nil, fmt.Errorf("smtp tls dial failed: %w", err)
		}
		client, err := smtp.NewClient(connection, config.Host)
		if err != nil {
			return nil, fmt.Errorf("smtp client init failed: %w", err)
		}
		return client, nil
	case "plain", "starttls":
		connection, err := dialer.Dial("tcp", address)
		if err != nil {
			return nil, fmt.Errorf("smtp dial failed: %w", err)
		}
		client, err := smtp.NewClient(connection, config.Host)
		if err != nil {
			return nil, fmt.Errorf("smtp client init failed: %w", err)
		}
		if config.Security == "starttls" {
			if ok, _ := client.Extension("STARTTLS"); !ok {
				client.Close()
				return nil, fmt.Errorf("smtp server does not advertise STARTTLS")
			}
			if err := client.StartTLS(&tls.Config{ServerName: config.Host}); err != nil {
				client.Close()
				return nil, fmt.Errorf("smtp STARTTLS failed: %w", err)
			}
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported security %q", config.Security)
	}
}

func authenticate(client *smtp.Client, config Config) error {
	if config.Auth == "none" {
		return nil
	}

	auth, err := authFor(config)
	if err != nil {
		return err
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}
	return nil
}

func lookupSecret(inline string, envName string, filePath string, kind string) (string, error) {
	switch {
	case inline != "":
		return inline, nil
	case envName != "":
		value, ok := os.LookupEnv(envName)
		if !ok || strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("%s env %s is empty or not set", kind, envName)
		}
		return value, nil
	case filePath != "":
		payload, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read %s file %s: %w", kind, filePath, err)
		}
		value := strings.TrimSpace(string(payload))
		if value == "" {
			return "", fmt.Errorf("%s file %s is empty", kind, filePath)
		}
		return value, nil
	default:
		return "", fmt.Errorf("%s is required for auth mode %s", kind, kindToAuth(kind))
	}
}

func kindToAuth(kind string) string {
	return "password"
}

func stdinHasData() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}
