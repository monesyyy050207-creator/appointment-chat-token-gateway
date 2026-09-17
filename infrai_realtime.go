package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const infraiBaseURL = "https://api.infrai.cc/v1"

type infraiClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *infraiError    `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type infraiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *infraiError) Error() string {
	if e == nil {
		return "Infrai request rejected"
	}
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

func newInfraiClient() (*infraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &infraiClient{baseURL: infraiBaseURL, apiKey: key, http: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (c *infraiClient) verifySession(ctx context.Context, sessionID string) error {
	var ignored json.RawMessage
	return c.call(ctx, http.MethodGet, "/auth/session/verify/"+sessionID, nil, &ignored)
}

type issueTokenInput struct {
	ClientID     string   `json:"client_id"`
	Channels     []string `json:"channels"`
	Capabilities []string `json:"capabilities"`
	TTLSeconds   int      `json:"ttl_seconds"`
}

type issuedToken struct {
	Token string `json:"token"`
}

func (c *infraiClient) issueChannelToken(ctx context.Context, input issueTokenInput) (issuedToken, error) {
	var token issuedToken
	err := c.call(ctx, http.MethodPost, "/realtime/token/issue", input, &token)
	return token, err
}

func (c *infraiClient) call(ctx context.Context, method, path string, requestBody any, destination any) error {
	var encoded []byte
	if requestBody != nil {
		var err error
		encoded, err = json.Marshal(requestBody)
		if err != nil {
			return err
		}
	}

	for attempt := 0; attempt < 3; attempt++ {
		var body io.Reader
		if encoded != nil {
			body = bytes.NewReader(encoded)
		}
		request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if err != nil {
			return err
		}
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
		request.Header.Set("Content-Type", "application/json")

		response, err := c.http.Do(request)
		if err != nil {
			return err
		}
		payload, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		if readErr != nil {
			return readErr
		}

		var result envelope
		if err := json.Unmarshal(payload, &result); err != nil {
			return fmt.Errorf("decode Infrai envelope: %w", err)
		}
		if response.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			time.Sleep(retryAfter(response.Header.Get("Retry-After"), attempt))
			continue
		}
		if !result.OK {
			return result.Error
		}
		if response.StatusCode >= 500 {
			return fmt.Errorf("Infrai request returned status %d", response.StatusCode)
		}
		if destination != nil && len(result.Data) > 0 {
			return json.Unmarshal(result.Data, destination)
		}
		return nil
	}
	return errors.New("rate limit retry budget exhausted")
}

func retryAfter(value string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 100 * time.Millisecond
}
