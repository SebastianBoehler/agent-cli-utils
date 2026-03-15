package smtpx

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestSendPlainNoAuth(t *testing.T) {
	server := newFakeSMTPServer(t)
	defer server.Close()

	result, err := Send(Config{
		Provider: "generic",
		Host:     "127.0.0.1",
		Port:     server.Port(),
		Security: "plain",
		Auth:     "none",
		From:     "sender@example.com",
	}, Message{
		To:      []string{"student@example.edu"},
		Subject: "Hello",
		Text:    "Body",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if result.Status != "sent" {
		t.Fatalf("Status = %q, want sent", result.Status)
	}
	if !strings.Contains(server.Data(), "Subject: Hello\r\n") {
		t.Fatalf("server payload missing subject: %q", server.Data())
	}
}

type fakeSMTPServer struct {
	listener net.Listener
	payload  strings.Builder
}

func newFakeSMTPServer(t *testing.T) *fakeSMTPServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	server := &fakeSMTPServer{listener: listener}
	go server.serve()
	return server
}

func (server *fakeSMTPServer) Port() int {
	return server.listener.Addr().(*net.TCPAddr).Port
}

func (server *fakeSMTPServer) Close() {
	_ = server.listener.Close()
}

func (server *fakeSMTPServer) Data() string {
	return server.payload.String()
}

func (server *fakeSMTPServer) serve() {
	connection, err := server.listener.Accept()
	if err != nil {
		return
	}
	defer connection.Close()

	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)
	writeLine(writer, "220 localhost ready")

	dataMode := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		if dataMode {
			if line == ".\r\n" {
				dataMode = false
				writeLine(writer, "250 ok")
				continue
			}
			server.payload.WriteString(line)
			continue
		}

		command := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(command, "EHLO"), strings.HasPrefix(command, "HELO"):
			writeLine(writer, "250-localhost")
			writeLine(writer, "250 OK")
		case strings.HasPrefix(command, "MAIL FROM"):
			writeLine(writer, "250 ok")
		case strings.HasPrefix(command, "RCPT TO"):
			writeLine(writer, "250 ok")
		case command == "DATA":
			writeLine(writer, "354 end with .")
			dataMode = true
		case command == "NOOP":
			writeLine(writer, "250 ok")
		case command == "QUIT":
			writeLine(writer, "221 bye")
			return
		default:
			writeLine(writer, fmt.Sprintf("250 ok %s", command))
		}
	}
}

func writeLine(writer *bufio.Writer, line string) {
	_, _ = writer.WriteString(line + "\r\n")
	_ = writer.Flush()
}
