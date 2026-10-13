package headless

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewClientDerivesEndpoints(t *testing.T) {
	c := NewClient("https://lordfarquaadthecreator--comfy-headless/", "us-east", "/tmp/wf")
	if c.SubmitURL != "https://lordfarquaadthecreator--comfy-headless-infer-async.us-east.modal.run" {
		t.Errorf("SubmitURL = %q", c.SubmitURL)
	}
	if c.StatusURL != "https://lordfarquaadthecreator--comfy-headless-status.us-east.modal.run" {
		t.Errorf("StatusURL = %q", c.StatusURL)
	}
}

func TestFillTypedAndEmbedded(t *testing.T) {
	template := map[string]any{
		"a": map[string]any{
			"seed":  "{{seed}}",
			"text":  "hello {{name}} and {{seed}}",
			"empty": "{{negative_prompt}}",
			"link":  []any{"1", 0},
		},
	}
	filled, err := Fill(template, map[string]any{
		"seed": int64(42), "name": "fahad", "negative_prompt": "",
	})
	if err != nil {
		t.Fatalf("Fill: %v", err)
	}
	got := filled.(map[string]any)["a"].(map[string]any)
	if got["seed"] != int64(42) {
		t.Errorf("seed = %#v, want int64(42)", got["seed"])
	}
	if got["text"] != "hello fahad and 42" {
		t.Errorf("text = %q", got["text"])
	}
	if got["empty"] != "" {
		t.Errorf("empty = %#v, want \"\"", got["empty"])
	}
}

func TestFillMissingTokenFails(t *testing.T) {
	_, err := Fill(map[string]any{"x": "{{seed}}"}, map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "seed") {
		t.Fatalf("want missing-token error, got %v", err)
	}
}

func TestFillRawTokenGuard(t *testing.T) {
	// "{{broken" does not match the token regex, so no "missing" is detected —
	// the belt-and-braces guard must still refuse to send it.
	_, err := Fill(map[string]any{"x": "{{broken"}, map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "raw {{token}}") {
		t.Fatalf("want raw-token error, got %v", err)
	}
}

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	template := map[string]any{
		"1": map[string]any{
			"class_type": "Test",
			"inputs": map[string]any{
				"positive": "{{positive_prompt}}",
				"seed":     "{{seed}}",
			},
		},
	}
	raw, _ := json.Marshal(template)
	if err := os.WriteFile(filepath.Join(dir, "test.workflow.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}

	return &Client{
		HTTP:         srv.Client(),
		SubmitURL:    srv.URL + "/infer-async",
		StatusURL:    srv.URL + "/status",
		WorkflowsDir: dir,
		PollInterval: 5 * time.Millisecond,
		RunTimeout:   5 * time.Second,
	}
}

func TestRunEndToEnd(t *testing.T) {
	var mu sync.Mutex
	var submitted map[string]any
	polls := 0

	imageA := []byte("PNG-A")
	imageB := []byte("PNG-B")

	handler := http.NewServeMux()
	handler.HandleFunc("/infer-async", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode submit body: %v", err)
		}
		mu.Lock()
		submitted = body
		mu.Unlock()
		fmt.Fprint(w, `{"call_id": "call-1", "status": "started"}`)
	})
	handler.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("call_id") != "call-1" {
			t.Errorf("bad call_id: %q", r.URL.Query().Get("call_id"))
		}
		mu.Lock()
		polls++
		n := polls
		mu.Unlock()
		if n < 2 {
			fmt.Fprint(w, `{"status": "pending"}`)
			return
		}
		payload := map[string]any{
			"status": "done",
			"result": map[string]any{
				"outputs": []map[string]string{
					{"filename": "a.png", "data_b64": base64.StdEncoding.EncodeToString(imageA)},
					{"filename": "b.png", "data_b64": base64.StdEncoding.EncodeToString(imageB)},
				},
				"texts": []string{},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	})

	client := newTestClient(t, handler)
	outputs, err := client.Run(context.Background(), "test.workflow.json",
		map[string]any{"positive_prompt": "a cat", "seed": int64(7)},
		map[string][]byte{"input.png": []byte("upload-me")})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(outputs) != 2 || string(outputs[0].Data) != "PNG-A" || outputs[1].Filename != "b.png" {
		t.Fatalf("outputs = %+v", outputs)
	}

	mu.Lock()
	defer mu.Unlock()
	// The submitted workflow must carry typed values and no raw tokens.
	raw, _ := json.Marshal(submitted["workflow"])
	if strings.Contains(string(raw), "{{") {
		t.Errorf("raw token in submitted workflow: %s", raw)
	}
	if !strings.Contains(string(raw), `"seed":7`) {
		t.Errorf("seed missing/typed wrong: %s", raw)
	}
	inputs := submitted["inputs"].(map[string]any)
	if inputs["input.png"] != base64.StdEncoding.EncodeToString([]byte("upload-me")) {
		t.Errorf("inputs = %#v", inputs)
	}
}

func TestRunSubmitError(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/infer-async", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"nope"}`, http.StatusBadRequest)
	})
	client := newTestClient(t, handler)
	_, err := client.Run(context.Background(), "test.workflow.json",
		map[string]any{"positive_prompt": "x", "seed": int64(1)}, nil)
	if err == nil || !strings.Contains(err.Error(), "HTTP 400") {
		t.Fatalf("want submit HTTP error, got %v", err)
	}
}

func TestRunExpired(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/infer-async", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"call_id": "call-9"}`)
	})
	handler.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status": "expired"}`)
	})
	client := newTestClient(t, handler)
	_, err := client.Run(context.Background(), "test.workflow.json",
		map[string]any{"positive_prompt": "x", "seed": int64(1)}, nil)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("want expired error, got %v", err)
	}
}

func TestRunJobError(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/infer-async", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"call_id": "call-3"}`)
	})
	handler.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status": "done", "result": {"error": {"node_errors": "missing model"}}}`)
	})
	client := newTestClient(t, handler)
	_, err := client.Run(context.Background(), "test.workflow.json",
		map[string]any{"positive_prompt": "x", "seed": int64(1)}, nil)
	if err == nil || !strings.Contains(err.Error(), "missing model") {
		t.Fatalf("want job error, got %v", err)
	}
}
