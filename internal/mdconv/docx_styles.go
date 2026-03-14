package mdconv

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func headingLevel(style string) int {
	lower := strings.ToLower(style)
	for i := 1; i <= 6; i++ {
		if strings.Contains(lower, fmt.Sprintf("heading %d", i)) || strings.Contains(lower, fmt.Sprintf("heading%d", i)) {
			return i
		}
	}
	return 0
}

func styleName(styles map[string]string, styleID string) string {
	if styleID == "" {
		return ""
	}
	if name, ok := styles[styleID]; ok {
		return name
	}
	return styleID
}

func docxStyles(reader *zip.Reader) (map[string]string, error) {
	data, err := zipEntry(reader, "word/styles.xml")
	if err != nil {
		return map[string]string{}, nil
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	styles := map[string]string{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return styles, nil
		}
		if err != nil {
			return nil, fmt.Errorf("parse docx styles: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "style" {
			continue
		}

		id := attrValue(start.Attr, "styleId")
		name := id
		for {
			token, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			switch typed := token.(type) {
			case xml.StartElement:
				if typed.Name.Local == "name" {
					if value := attrValue(typed.Attr, "val"); value != "" {
						name = value
					}
				}
			case xml.EndElement:
				if typed.Name.Local == start.Name.Local {
					if id != "" {
						styles[id] = name
					}
					goto nextStyle
				}
			}
		}
	nextStyle:
	}
}
