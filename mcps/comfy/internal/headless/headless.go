// Package headless talks to the comfy-headless Modal app: it fills a
// workflow template from the shared store with typed values, uploads input
// images, submits the job to the async endpoint, polls until it completes,
// and returns the decoded outputs.
//
// This replaces the per-worker HTTP clients: comfy-headless runs every
// fleet workflow natively (see senor-modal-apps/comfy-headless/workflows).
package headless

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	DefaultPollInterval = 3 * time.Second
	DefaultRunTimeout   = 30 * time.Minute
	DefaultHTTPTimeout  = 120 * time.Second
)

var (
	tokenRe      = regexp.MustCompile(`\{\{(\w+)\}\}`)
	exactTokenRe = regexp.MustCompile(`^\{\{(\w+)\}\}$`)
)

// Output is one file produced by the workflow (PNG, mp4, ...).
type Output struct {
	Filename string
	Data     []byte
}

// Client submits workflows to one comfy-headless deployment.
type Client struct {
	HTTP         *http.Client
	SubmitURL    string // POST {workflow, inputs} -> {call_id}
	StatusURL    string // GET ?call_id= -> {status, result}
	WorkflowsDir string // directory holding the *.workflow.json store

	PollInterval time.Duration
	RunTimeout   time.Duration
}

// NewClient derives the two endpoint URLs from the comfy-headless base
// (e.g. https://workspace--comfy-headless) and the Modal region.
func NewClient(baseURL, region, workflowsDir string) *Client {
	base := strings.TrimRight(baseURL, "/")
	return &Client{
		HTTP:         &http.Client{Timeout: DefaultHTTPTimeout},
		SubmitURL:    fmt.Sprintf("%s-infer-async.%s.modal.run", base, region),
		StatusURL:    fmt.Sprintf("%s-status.%s.modal.run", base, region),
		WorkflowsDir: workflowsDir,
		PollInterval: DefaultPollInterval,
		RunTimeout:   DefaultRunTimeout,
	}
}

// Run executes one workflow end to end. values must cover every {{token}} the
// template contains (an empty string is a legal value); images are keyed by
// the LoadImage filenames the workflow references.
func (c *Client) Run(ctx context.Context, workflowFile string, values map[string]any, images map[string][]byte) ([]Output, error) {
	template, err := c.loadTemplate(workflowFile)
	if err != nil {
		return nil, err
	}

	filled, err := Fill(template, values)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", workflowFile, err)
	}

	inputs := make(map[string]string, len(images))
	for name, data := range images {
		inputs[name] = base64.StdEncoding.EncodeToString(data)
	}

	body, err := json.Marshal(map[string]any{"workflow": filled, "inputs": inputs})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	callID, err := c.submit(ctx, body)
	if err != nil {
		return nil, err
	}
	return c.poll(ctx, callID)
}

func (c *Client) loadTemplate(workflowFile string) (any, error) {
	path := filepath.Join(c.WorkflowsDir, workflowFile)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow %s: %w", path, err)
	}
	var template any
	if err := json.Unmarshal(raw, &template); err != nil {
		return nil, fmt.Errorf("parse workflow %s: %w", path, err)
	}
	return template, nil
}

// Fill replaces every {{token}} in the template with its value. A token with
// no value, or a raw token surviving the fill, is an error — nothing partial
// ever leaves this function.
func Fill(template any, values map[string]any) (any, error) {
	missing := map[string]bool{}
	filled := fillValue(template, values, missing)
	if len(missing) > 0 {
		names := make([]string, 0, len(missing))
		for name := range missing {
			names = append(names, name)
		}
		return nil, fmt.Errorf("missing values for tokens: %s", strings.Join(names, ", "))
	}
	raw, err := json.Marshal(filled)
	if err != nil {
		return nil, fmt.Errorf("marshal filled workflow: %w", err)
	}
	if bytes.Contains(raw, []byte("{{")) {
		return nil, fmt.Errorf("internal error: a raw {{token}} survived filling")
	}
	return filled, nil
}

func fillValue(v any, values map[string]any, missing map[string]bool) any {
	switch t := v.(type) {
	case string:
		if m := exactTokenRe.FindStringSubmatch(t); m != nil {
			val, ok := values[m[1]]
			if !ok {
				missing[m[1]] = true
				return t
			}
			return val
		}
		return tokenRe.ReplaceAllStringFunc(t, func(tok string) string {
			name := tokenRe.FindStringSubmatch(tok)[1]
			val, ok := values[name]
			if !ok {
				missing[name] = true
				return tok
			}
			return fmt.Sprint(val)
		})
	case map[string]any:
		for k, vv := range t {
			t[k] = fillValue(vv, values, missing)
		}
		return t
	case []any:
		for i, vv := range t {
			t[i] = fillValue(vv, values, missing)
		}
		return t
	default:
		return v
	}
}

func (c *Client) submit(ctx context.Context, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.SubmitURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build submit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("submit to %s: %w", c.SubmitURL, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("submit HTTP %d: %s", resp.StatusCode, truncate(data))
	}
	var out struct {
		CallID string `json:"call_id"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("parse submit response: %w (%s)", err, truncate(data))
	}
	if out.CallID == "" {
		return "", fmt.Errorf("empty call_id in submit response: %s", truncate(data))
	}
	return out.CallID, nil
}

type statusResponse struct {
	Status string          `json:"status"`
	Result json.RawMessage `json:"result"`
}

type jobResult struct {
	Outputs []struct {
		Filename string `json:"filename"`
		DataB64  string `json:"data_b64"`
	} `json:"outputs"`
	Texts []string        `json:"texts"`
	Error json.RawMessage `json:"error"`
}

func (c *Client) poll(ctx context.Context, callID string) ([]Output, error) {
	deadline := time.Now().Add(c.RunTimeout)

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("job %s still pending after %s — stopped polling", callID, c.RunTimeout)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.PollInterval):
		}

		status, err := c.pollOnce(ctx, callID)
		if err != nil {
			return nil, err
		}

		switch status.Status {
		case "pending":
			continue
		case "done":
			return decodeOutputs(status.Result)
		case "expired":
			return nil, fmt.Errorf("job %s result expired", callID)
		case "bad_call_id":
			return nil, fmt.Errorf("comfy-headless rejected call_id %q", callID)
		default:
			return nil, fmt.Errorf("unexpected status %q for job %s: %s", status.Status, callID, truncate(status.Result))
		}
	}
}

func (c *Client) pollOnce(ctx context.Context, callID string) (*statusResponse, error) {
	endpoint := c.StatusURL + "?call_id=" + url.QueryEscape(callID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build status request: %w", err)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("poll %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status HTTP %d: %s", resp.StatusCode, truncate(data))
	}
	var status statusResponse
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("parse status response: %w (%s)", err, truncate(data))
	}
	return &status, nil
}

func decodeOutputs(raw json.RawMessage) ([]Output, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("done status carried no result")
	}
	var result jobResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse job result: %w (%s)", err, truncate(raw))
	}
	if len(result.Error) > 0 && !bytes.Equal(result.Error, []byte("null")) {
		return nil, fmt.Errorf("job failed: %s", truncate(result.Error))
	}
	if len(result.Outputs) == 0 {
		return nil, fmt.Errorf("job produced no outputs (texts: %v)", result.Texts)
	}

	outputs := make([]Output, 0, len(result.Outputs))
	for _, o := range result.Outputs {
		data, err := base64.StdEncoding.DecodeString(o.DataB64)
		if err != nil {
			return nil, fmt.Errorf("decode output %q: %w", o.Filename, err)
		}
		outputs = append(outputs, Output{Filename: o.Filename, Data: data})
	}
	return outputs, nil
}

func truncate(data []byte) string {
	const limit = 500
	if len(data) <= limit {
		return string(data)
	}
	return string(data[:limit]) + "…"
}
