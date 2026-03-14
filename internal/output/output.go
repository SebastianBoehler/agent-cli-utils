package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func Write(format string, value any) error {
	return WriteTo(os.Stdout, format, value)
}

func WriteTo(writer io.Writer, format string, value any) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case "yaml":
		payload, err := yaml.Marshal(value)
		if err != nil {
			return err
		}

		if _, err := writer.Write(payload); err != nil {
			return err
		}

		if len(payload) == 0 || payload[len(payload)-1] != '\n' {
			_, err = fmt.Fprintln(writer)
			return err
		}

		return nil
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}
