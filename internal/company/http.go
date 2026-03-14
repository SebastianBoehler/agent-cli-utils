package company

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	neturl "net/url"
	"strings"
	"time"
)

func cloneHTTPClient(client *http.Client, timeout time.Duration) *http.Client {
	if client == nil {
		client = &http.Client{}
	}

	cloned := *client
	if cloned.Timeout <= 0 {
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		cloned.Timeout = timeout
	}
	return &cloned
}

func newCookieClient(client *http.Client, timeout time.Duration) (*http.Client, error) {
	cloned := cloneHTTPClient(client, timeout)
	if cloned.Jar != nil {
		return cloned, nil
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	cloned.Jar = jar
	return cloned, nil
}

func readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected http %d: %s", resp.StatusCode, trimSnippet(string(payload)))
	}
	return payload, nil
}

func trimSnippet(text string) string {
	text = strings.TrimSpace(text)
	if len(text) > 160 {
		return text[:160]
	}
	if text == "" {
		return http.StatusText(http.StatusBadGateway)
	}
	return text
}

func resolveURL(base string, href string) string {
	if strings.TrimSpace(href) == "" {
		return ""
	}

	parsedBase, err := neturl.Parse(base)
	if err != nil {
		return href
	}
	parsedRef, err := neturl.Parse(href)
	if err != nil {
		return href
	}
	return parsedBase.ResolveReference(parsedRef).String()
}
