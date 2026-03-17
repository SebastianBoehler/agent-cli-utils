package samsungcontrol

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/devicestore"
	"github.com/gorilla/websocket"
)

type Service struct{}

type message struct {
	Event string `json:"event"`
	Data  struct {
		Token string `json:"token"`
	} `json:"data"`
}

func NewService() *Service {
	return &Service{}
}

func (service *Service) Pair(ctx context.Context, options PairOptions) (PairResult, error) {
	ctx, cancel := withTimeout(ctx, options.Timeout)
	defer cancel()

	host := strings.TrimSpace(options.Host)
	if host == "" {
		return PairResult{}, fmt.Errorf("host is required")
	}
	token, err := service.connect(ctx, host, firstNonEmpty(options.Name, "agentsamsung"), "")
	if err != nil {
		return PairResult{}, err
	}
	if err := devicestore.SaveSamsungToken(host, token); err != nil {
		return PairResult{}, err
	}
	return PairResult{Protocol: Protocol, Target: host, OK: true, Detail: "pairing accepted and token stored", Token: token}, nil
}

func (service *Service) Remote(ctx context.Context, options RemoteOptions) (ActionResult, error) {
	ctx, cancel := withTimeout(ctx, options.Timeout)
	defer cancel()

	conn, err := service.authorizedDial(ctx, strings.TrimSpace(options.Host))
	if err != nil {
		return ActionResult{}, err
	}
	defer conn.Close()

	payload := map[string]any{"method": "ms.remote.control", "params": map[string]any{"Cmd": "Click", "DataOfCmd": options.Key, "Option": "false", "TypeOfRemote": "SendRemoteKey"}}
	if err := conn.WriteJSON(payload); err != nil {
		return ActionResult{}, fmt.Errorf("send samsung key: %w", err)
	}
	return ActionResult{Operation: "remote", Protocol: Protocol, Target: options.Host, OK: true, Detail: "key sent"}, nil
}

func (service *Service) Launch(ctx context.Context, options LaunchOptions) (ActionResult, error) {
	ctx, cancel := withTimeout(ctx, options.Timeout)
	defer cancel()

	conn, err := service.authorizedDial(ctx, strings.TrimSpace(options.Host))
	if err != nil {
		return ActionResult{}, err
	}
	defer conn.Close()

	payload := map[string]any{"method": "ms.channel.emit", "params": map[string]any{"event": "ed.apps.launch", "to": "host", "data": map[string]any{"action_type": firstNonEmpty(options.AppType, "DEEP_LINK"), "appId": options.AppID, "metaTag": options.MetaTag}}}
	if err := conn.WriteJSON(payload); err != nil {
		return ActionResult{}, fmt.Errorf("launch samsung app: %w", err)
	}
	return ActionResult{Operation: "launch", Protocol: Protocol, Target: options.Host, OK: true, Detail: "launch command sent"}, nil
}

func (service *Service) authorizedDial(ctx context.Context, host string) (*websocket.Conn, error) {
	token, err := devicestore.LoadSamsungToken(host)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("no samsung token stored for %s; run pair first", host)
	}
	conn, _, err := service.dial(ctx, host, "agentsamsung", token)
	return conn, err
}

func (service *Service) connect(ctx context.Context, host string, name string, token string) (string, error) {
	conn, msg, err := service.dial(ctx, host, name, token)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if token == "" && strings.TrimSpace(msg.Data.Token) == "" {
		return "", fmt.Errorf("pairing connected but no token was returned")
	}
	if msg.Data.Token != "" {
		return msg.Data.Token, nil
	}
	return token, nil
}

func (service *Service) dial(ctx context.Context, host string, name string, token string) (*websocket.Conn, message, error) {
	rawURL := fmt.Sprintf("wss://%s:8002/api/v2/channels/samsung.remote.control?name=%s", host, url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(name))))
	if strings.TrimSpace(token) != "" {
		rawURL += "&token=" + url.QueryEscape(token)
	}
	dialer := websocket.Dialer{
		Proxy:            websocket.DefaultDialer.Proxy,
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
	}
	conn, _, err := dialer.DialContext(ctx, rawURL, nil)
	if err != nil {
		return nil, message{}, fmt.Errorf("connect samsung remote: %w", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var msg message
	if err := conn.ReadJSON(&msg); err != nil {
		conn.Close()
		return nil, message{}, fmt.Errorf("read samsung handshake: %w", err)
	}
	if msg.Event != "ms.channel.connect" {
		conn.Close()
		return nil, msg, fmt.Errorf("samsung pairing failed: %s", firstNonEmpty(msg.Event, "unexpected response"))
	}
	return conn, msg, nil
}

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return context.WithTimeout(ctx, timeout)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
