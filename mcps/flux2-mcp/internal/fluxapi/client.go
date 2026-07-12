package fluxapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type GenerateRequest struct {
	Prompt            string     `json:"prompt"`
	ReferenceImages   []string   `json:"reference_images,omitempty"`
	Loras             []LoraSpec `json:"loras,omitempty"`
	Width             int        `json:"width"`
	Height            int        `json:"height"`
	NumInferenceSteps int        `json:"num_inference_steps"`
	GuidanceScale     float64    `json:"guidance_scale"`
	Seed              *int64     `json:"seed,omitempty"`
	NegativePrompt    string     `json:"negative_prompt,omitempty"`
	Variant           string     `json:"variant,omitempty"`
}

type LoraSpec struct {
	Path     string  `json:"path"`
	Strength float64 `json:"strength"`
}

type AsyncResponse struct {
	CallID string `json:"call_id"`
}

type StatusResponse struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

const (
	DefaultTimeout  = 300 * time.Second
	PollInterval    = 1500 * time.Millisecond
	MaxPollAttempts = 200
)

type Client struct {
	HTTP   *http.Client
	APIURL string
}

func NewClient(apiURL string) *Client {
	return &Client{
		HTTP:   &http.Client{Timeout: DefaultTimeout},
		APIURL: strings.TrimRight(apiURL, "/"),
	}
}

// Generate sends an async request then polls until the PNG is ready.
func (c *Client) Generate(ctx context.Context, req GenerateRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	log.Printf("[flux2-api] POST %s/api/v1/generate_async (%d bytes)", c.APIURL, len(body))
	t0 := time.Now()

	resp, err := c.HTTP.Post(c.APIURL+"/api/v1/generate_async", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("async request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(msg))
	}

	var asyncResp AsyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&asyncResp); err != nil {
		return nil, fmt.Errorf("decode async response: %w", err)
	}

	if asyncResp.CallID == "" {
		return nil, fmt.Errorf("empty call_id in response")
	}

	log.Printf("[flux2-api] got call_id=%s in %v", asyncResp.CallID, time.Since(t0))

	data, err := c.pollStatus(ctx, asyncResp.CallID)
	if err != nil {
		return nil, err
	}
	log.Printf("[flux2-api] total time %v, received %d bytes", time.Since(t0), len(data))
	return data, nil
}

func (c *Client) pollStatus(ctx context.Context, callID string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/api/v1/status/%s", c.APIURL, callID)

	for attempt := 1; attempt <= MaxPollAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(PollInterval):
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("build poll request: %w", err)
		}

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return nil, fmt.Errorf("poll request failed: %w", err)
		}

		ct := resp.Header.Get("Content-Type")
		if ct == "image/png" {
			data, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("read png response: %w", err)
			}
			log.Printf("[flux2-api] poll #%d: received image/png (%d bytes)", attempt, len(data))
			return data, nil
		}

		var status StatusResponse
		_ = json.NewDecoder(resp.Body).Decode(&status)
		resp.Body.Close()

		log.Printf("[flux2-api] poll #%d: status=%s", attempt, status.Status)

		switch status.Status {
		case "pending":
			continue
		case "expired":
			return nil, fmt.Errorf("generation result expired")
		case "error":
			return nil, fmt.Errorf("generation error: %s", status.Detail)
		}
	}

	return nil, fmt.Errorf("polling timed out after %d attempts", MaxPollAttempts)
}
