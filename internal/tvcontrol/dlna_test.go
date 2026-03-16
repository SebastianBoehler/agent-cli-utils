package tvcontrol

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlayDLNA(t *testing.T) {
	requests := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		payload, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		requests = append(requests, request.Header.Get("SOAPAction")+" "+string(payload))
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("<ok/>"))
	}))
	defer server.Close()

	result, err := playDLNA(context.Background(), server.Client(), Device{
		Name:       "TV",
		Protocol:   ProtocolDLNA,
		ControlURL: server.URL,
	}, "http://example.test/stream.m3u8")
	if err != nil {
		t.Fatalf("play dlna: %v", err)
	}

	if !result.OK {
		t.Fatalf("expected OK result: %+v", result)
	}
	if len(requests) != 2 {
		t.Fatalf("expected 2 SOAP requests, got %d", len(requests))
	}
	if !strings.Contains(requests[0], "#SetAVTransportURI") || !strings.Contains(requests[0], "<CurrentURI>http://example.test/stream.m3u8</CurrentURI>") {
		t.Fatalf("unexpected first request %q", requests[0])
	}
	if !strings.Contains(requests[1], "#Play") || !strings.Contains(requests[1], "<Speed>1</Speed>") {
		t.Fatalf("unexpected second request %q", requests[1])
	}
}

func TestResolveControlURL(t *testing.T) {
	result, err := resolveControlURL(mustParseURL(t, "http://192.168.1.10:8080/desc.xml"), "", "/upnp/control/avtransport1")
	if err != nil {
		t.Fatalf("resolve control url: %v", err)
	}
	if result != "http://192.168.1.10:8080/upnp/control/avtransport1" {
		t.Fatalf("unexpected resolved url %q", result)
	}
}
