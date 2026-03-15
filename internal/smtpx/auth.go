package smtpx

import (
	"fmt"
	"net/smtp"
)

func secretFor(config Config) (string, error) {
	switch config.Auth {
	case "none":
		return "", nil
	case "password":
		return lookupSecret(config.Password, config.PasswordEnv, config.PasswordFile, "password")
	default:
		return "", fmt.Errorf("unsupported auth %q", config.Auth)
	}
}

func authFor(config Config) (smtp.Auth, error) {
	secret, err := secretFor(config)
	if err != nil {
		return nil, err
	}

	switch config.Auth {
	case "none":
		return nil, nil
	case "password":
		return smtp.PlainAuth("", config.Username, secret, config.Host), nil
	default:
		return nil, fmt.Errorf("unsupported auth %q", config.Auth)
	}
}
