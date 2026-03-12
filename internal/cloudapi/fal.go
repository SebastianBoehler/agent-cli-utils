package cloudapi

import (
	"fmt"
	"net/http"
	"strings"
)

const defaultFalBaseURL = "https://queue.fal.run"

type FalClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
	baseURL    string
}

func NewFalClient(httpClient *http.Client, apiKey string, model string, baseURL string) (*FalClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("missing fal API key")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("missing fal model path")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultFalBaseURL
	}

	return &FalClient{
		httpClient: httpClient,
		apiKey:     strings.TrimSpace(apiKey),
		model:      strings.TrimSpace(model),
		baseURL:    strings.TrimSpace(baseURL),
	}, nil
}

func (client *FalClient) Submit(payload any) (Result, error) {
	body := payload
	if body == nil {
		body = map[string]any{}
	}

	endpoint := JoinURL(client.baseURL, client.model)
	statusCode, data, err := DoJSON(client.httpClient, http.MethodPost, endpoint, map[string]string{
		"Authorization": falAuthHeader(client.apiKey),
	}, body)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Service:    "fal",
		Operation:  "submit",
		Target:     client.model,
		RequestID:  ExtractString(data, "request_id"),
		Status:     ExtractString(data, "status"),
		HTTPStatus: statusCode,
		URL:        endpoint,
		Data:       data,
	}
	if statusCode >= http.StatusBadRequest {
		result.Error = ExtractError(data)
	}
	return result, nil
}

func (client *FalClient) Status(requestID string, includeLogs bool) (Result, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return Result{}, fmt.Errorf("missing request ID")
	}

	endpoint := JoinURL(client.baseURL, client.model, "requests", requestID, "status")
	var err error
	if includeLogs {
		endpoint, err = AddQuery(endpoint, map[string]string{"logs": "1"})
		if err != nil {
			return Result{}, err
		}
	}

	statusCode, data, err := DoJSON(client.httpClient, http.MethodGet, endpoint, map[string]string{
		"Authorization": falAuthHeader(client.apiKey),
	}, nil)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Service:    "fal",
		Operation:  "status",
		Target:     client.model,
		RequestID:  requestID,
		Status:     ExtractString(data, "status"),
		HTTPStatus: statusCode,
		URL:        endpoint,
		Data:       data,
	}
	if statusCode >= http.StatusBadRequest {
		result.Error = ExtractError(data)
	}
	return result, nil
}

func (client *FalClient) Result(requestID string) (Result, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return Result{}, fmt.Errorf("missing request ID")
	}

	endpoint := JoinURL(client.baseURL, client.model, "requests", requestID)
	statusCode, data, err := DoJSON(client.httpClient, http.MethodGet, endpoint, map[string]string{
		"Authorization": falAuthHeader(client.apiKey),
	}, nil)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Service:    "fal",
		Operation:  "result",
		Target:     client.model,
		RequestID:  requestID,
		Status:     ExtractString(data, "status"),
		HTTPStatus: statusCode,
		URL:        endpoint,
		Data:       data,
	}
	if statusCode >= http.StatusBadRequest {
		result.Error = ExtractError(data)
	}
	return result, nil
}

func (client *FalClient) Cancel(requestID string) (Result, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return Result{}, fmt.Errorf("missing request ID")
	}

	endpoint := JoinURL(client.baseURL, client.model, "requests", requestID, "cancel")
	statusCode, data, err := DoJSON(client.httpClient, http.MethodPut, endpoint, map[string]string{
		"Authorization": falAuthHeader(client.apiKey),
	}, nil)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Service:    "fal",
		Operation:  "cancel",
		Target:     client.model,
		RequestID:  requestID,
		Status:     ExtractString(data, "status"),
		HTTPStatus: statusCode,
		URL:        endpoint,
		Data:       data,
	}
	if statusCode >= http.StatusBadRequest {
		result.Error = ExtractError(data)
	}
	return result, nil
}

func falAuthHeader(apiKey string) string {
	key := strings.TrimSpace(apiKey)
	if strings.Contains(key, " ") {
		return key
	}
	return "Key " + key
}
