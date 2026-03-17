package tvcontrol

import (
	_ "embed"
	"os"
)

//go:embed agenttv_airplay_helper.py
var appleAirPlayHelperSource string

func writeAppleHelper() (string, error) {
	file, err := os.CreateTemp("", "agenttv-airplay-*.py")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.WriteString(appleAirPlayHelperSource); err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	if err := file.Chmod(0o700); err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}
