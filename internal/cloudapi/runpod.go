package cloudapi

import (
	"fmt"
	"net/http"
	"strings"
)

const defaultRunPodBaseURL = "https://api.runpod.ai/v2"

type RunPodClient struct {
	httpClient *http.Client
	apiKey     string
	endpointID string
	baseURL    string
}

func NewRunPodClient(httpClient *http.Client, apiKey string, endpointID string, baseURL string) (*RunPodClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("missing RunPod API key")
	}
	if strings.TrimSpace(endpointID) == "" {
		return nil, fmt.Errorf("missing RunPod endpoint ID")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultRunPodBaseURL
	}

	return &RunPodClient{
		httpClient: httpClient,
		apiKey:     strings.TrimSpace(apiKey),
		endpointID: strings.TrimSpace(endpointID),
		baseURL:    strings.TrimSpace(baseURL),
	}, nil
}

func (client *RunPodClient) Submit(payload any, synchronous bool) (Result, error) {
	body := payload
	if body == nil {
		body = map[string]any{}
	}

	mode := "async"
	endpoint := JoinURL(client.baseURL, client.endpointID, "run")
	if synchronous {
		mode = "sync"
		endpoint = JoinURL(client.baseURL, client.endpointID, "runsync")
	}

	statusCode, data, err := DoJSON(client.httpClient, http.MethodPost, endpoint, map[string]string{
		"Authorization": runPodAuthHeader(client.apiKey),
	}, body)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Service:    "runpod",
		Operation:  "submit",
		Mode:       mode,
		Target:     client.endpointID,
		RequestID:  ExtractString(data, "id", "request_id"),
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

func (client *RunPodClient) Status(requestID string) (Result, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return Result{}, fmt.Errorf("missing request ID")
	}

	endpoint := JoinURL(client.baseURL, client.endpointID, "status", requestID)
	statusCode, data, err := DoJSON(client.httpClient, http.MethodGet, endpoint, map[string]string{
		"Authorization": runPodAuthHeader(client.apiKey),
	}, nil)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Service:    "runpod",
		Operation:  "status",
		Target:     client.endpointID,
		RequestID:  ExtractString(data, "id", "request_id"),
		Status:     ExtractString(data, "status"),
		HTTPStatus: statusCode,
		URL:        endpoint,
		Data:       data,
	}
	if result.RequestID == "" {
		result.RequestID = requestID
	}
	if statusCode >= http.StatusBadRequest {
		result.Error = ExtractError(data)
	}
	return result, nil
}

func (client *RunPodClient) Result(requestID string) (Result, error) {
	result, err := client.Status(requestID)
	if err != nil {
		return Result{}, err
	}
	result.Operation = "result"
	return result, nil
}

func (client *RunPodClient) Cancel(requestID string) (Result, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return Result{}, fmt.Errorf("missing request ID")
	}

	endpoint := JoinURL(client.baseURL, client.endpointID, "cancel", requestID)
	statusCode, data, err := DoJSON(client.httpClient, http.MethodPost, endpoint, map[string]string{
		"Authorization": runPodAuthHeader(client.apiKey),
	}, nil)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Service:    "runpod",
		Operation:  "cancel",
		Target:     client.endpointID,
		RequestID:  ExtractString(data, "id", "request_id"),
		Status:     ExtractString(data, "status"),
		HTTPStatus: statusCode,
		URL:        endpoint,
		Data:       data,
	}
	if result.RequestID == "" {
		result.RequestID = requestID
	}
	if statusCode >= http.StatusBadRequest {
		result.Error = ExtractError(data)
	}
	return result, nil
}

func runPodAuthHeader(apiKey string) string {
	key := strings.TrimSpace(apiKey)
	if strings.Contains(key, " ") {
		return key
	}
	return "Bearer " + key
}
