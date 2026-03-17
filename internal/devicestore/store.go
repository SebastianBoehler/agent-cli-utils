package devicestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type credentialStore struct {
	AppleAirPlay map[string]string `json:"apple_airplay,omitempty"`
	SamsungToken map[string]string `json:"samsung_token,omitempty"`
}

func LoadAppleCredentials(host string) (string, error) {
	store, err := loadStore()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(store.AppleAirPlay[strings.TrimSpace(host)]), nil
}

func SaveAppleCredentials(host string, credentials string) error {
	store, err := loadStore()
	if err != nil {
		return err
	}
	store.AppleAirPlay[strings.TrimSpace(host)] = strings.TrimSpace(credentials)
	return saveStore(store)
}

func LoadSamsungToken(host string) (string, error) {
	store, err := loadStore()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(store.SamsungToken[strings.TrimSpace(host)]), nil
}

func SaveSamsungToken(host string, token string) error {
	store, err := loadStore()
	if err != nil {
		return err
	}
	store.SamsungToken[strings.TrimSpace(host)] = strings.TrimSpace(token)
	return saveStore(store)
}

func loadStore() (credentialStore, error) {
	path, err := storePath()
	if err != nil {
		return credentialStore{}, err
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return credentialStore{
				AppleAirPlay: map[string]string{},
				SamsungToken: map[string]string{},
			}, nil
		}
		return credentialStore{}, fmt.Errorf("read credential store: %w", err)
	}
	var store credentialStore
	if err := json.Unmarshal(payload, &store); err != nil {
		return credentialStore{}, fmt.Errorf("parse credential store: %w", err)
	}
	if store.AppleAirPlay == nil {
		store.AppleAirPlay = map[string]string{}
	}
	if store.SamsungToken == nil {
		store.SamsungToken = map[string]string{}
	}
	return store, nil
}

func saveStore(store credentialStore) error {
	path, err := storePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}
	payload, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credential store: %w", err)
	}
	if err := os.WriteFile(path, append(payload, '\n'), 0o600); err != nil {
		return fmt.Errorf("write credential store: %w", err)
	}
	return nil
}

func storePath() (string, error) {
	if custom := strings.TrimSpace(os.Getenv("AGENTTV_STORE")); custom != "" {
		return custom, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config", "agenttv", "credentials.json"), nil
}
