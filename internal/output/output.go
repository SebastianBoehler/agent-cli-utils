package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Write(format string, value any) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case "yaml":
		payload, err := yaml.Marshal(value)
		if err != nil {
			return err
		}

		if _, err := os.Stdout.Write(payload); err != nil {
			return err
		}

		if len(payload) == 0 || payload[len(payload)-1] != '\n' {
			_, err = fmt.Fprintln(os.Stdout)
			return err
		}

		return nil
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}
