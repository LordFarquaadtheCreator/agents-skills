package server

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

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcps/comfy/internal/headless"
	"mcps/comfy/internal/store"
)

const zitFixture = `{
 "1": {"class_type":"CheckpointLoaderSimple","inputs":{"ckpt_name":"{{unet_name}}"}},
 "2": {"class_type":"CLIPTextEncode","inputs":{"text":"{{positive_prompt}}"}},
 "3": {"class_type":"CLIPTextEncode","inputs":{"text":"{{negative_prompt}}"}},
 "4": {"class_type":"KSampler","inputs":{"seed":"{{seed}}","steps":"{{steps}}","cfg":"{{cfg}}","denoise":"{{denoise}}"}},
 "5": {"class_type":"LoraLoader","inputs":{"lora_name":"{{lora_name_1}}","strength_model":"{{lora_strength_1}}"}},
 "6": {"class_type":"EmptyLatentImage","inputs":{"width":"{{width}}","height":"{{height}}"}},
 "7": {"class_type":"SaveImage","inputs":{"filename_prefix":"prefix-{{seed}}","images":["4",0]}}
}`

const inpaintFixture = `{
 "1": {"class_type":"LoadImage","inputs":{"image":"mask.png"}},
 "2": {"class_type":"LoadImage","inputs":{"image":"mask.png"}},
 "3": {"class_type":"InpaintCropAndStitch","inputs":{"crop_width":"{{crop_width}}","crop_height":"{{crop_height}}","prompt":"{{positive_prompt}}"}},
 "4": {"class_type":"SaveImage","inputs":{"filename_prefix":"out","images":["3",0]}}
}`

type rig struct {
	session    *mcp.ClientSession
	submitted  func() map[string]any
	outputsDir string
}

// newRig wires: fixture store -> real store.Load -> server.Build over an
// in-memory transport -> mock comfy-headless. Everything is hermetic.
func newRig(t *testing.T, outputs map[string][]byte) *rig {
	t.Helper()

	storeDir := t.TempDir()
	for name, content := range map[string]string{
		"z-image-turbo-txt2img.workflow.json": zitFixture,
		"flux2-klein-9b-inpaint.workflow.json": inpaintFixture,
	} {
		if err := os.WriteFile(filepath.Join(storeDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	workflows, err := store.Load(storeDir)
	if err != nil {
		t.Fatalf("store.Load: %v", err)
	}

	var mu sync.Mutex
	var captured map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("/infer-async", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		captured = body
		mu.Unlock()
		fmt.Fprint(w, `{"call_id":"call-comfy-1"}`)
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]any{"status": "done"}
		// deterministic output order
		names := make([]string, 0, len(outputs))
		for name := range outputs {
			names = append(names, name)
		}
		sortStrings(names)
		outs := make([]map[string]string, 0, len(names))
		for _, name := range names {
			outs = append(outs, map[string]string{
				"filename": name,
				"data_b64": base64.StdEncoding.EncodeToString(outputs[name]),
			})
		}
		payload["result"] = map[string]any{"outputs": outs, "texts": []string{}}
		_ = json.NewEncoder(w).Encode(payload)
	})
	headlessSrv := httptest.NewServer(mux)
	t.Cleanup(headlessSrv.Close)

	client := &headless.Client{
		HTTP:         headlessSrv.Client(),
		SubmitURL:    headlessSrv.URL + "/infer-async",
		StatusURL:    headlessSrv.URL + "/status",
		WorkflowsDir: storeDir,
		PollInterval: 5 * time.Millisecond,
		RunTimeout:   5 * time.Second,
	}

	outputsDir := t.TempDir()
	server := Build(client, workflows, outputsDir)

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() { done <- server.Run(ctx, serverTransport) }()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	return &rig{
		session: session,
		submitted: func() map[string]any {
			mu.Lock()
			defer mu.Unlock()
			return captured
		},
		outputsDir: outputsDir,
	}
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	if args == nil {
		args = map[string]any{}
	}
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	return res
}

func textOf(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("no content")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content[0] is %T, want TextContent", res.Content[0])
	}
	return tc.Text
}

func validZitArgs() map[string]any {
	return map[string]any{
		"positive_prompt": "a test render",
		"unet_name":       "z_image_turbo_bf16.safetensors",
		"seed":            7,
		"steps":           16,
		"cfg":             1.0,
		"denoise":         1.0,
		"width":           512,
		"height":          512,
		"lora_name_1":     "style.safetensors",
		"lora_strength_1": 0.8,
	}
}

// --- tool listing / schema generation ---

func TestListTools_SchemasGeneratedFromStore(t *testing.T) {
	r := newRig(t, map[string][]byte{"out.png": []byte("png")})

	res, err := r.session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(res.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(res.Tools))
	}

	byName := map[string]*mcp.Tool{}
	for _, tool := range res.Tools {
		byName[tool.Name] = tool
	}
	zit, ok := byName["z_image_turbo_txt2img"]
	if !ok {
		t.Fatalf("z_image_turbo_txt2img tool missing: %v", byName)
	}

	schema, ok := zit.InputSchema.(map[string]any)
	if !ok {
		t.Fatalf("schema is %T", zit.InputSchema)
	}
	props, _ := schema["properties"].(map[string]any)
	if props == nil {
		t.Fatal("no properties")
	}
	// token types
	if p, _ := props["seed"].(map[string]any); p["type"] != "number" {
		t.Errorf("seed schema = %v", p)
	}
	if p, _ := props["positive_prompt"].(map[string]any); p["type"] != "string" {
		t.Errorf("positive_prompt schema = %v", p)
	}
	// required: everything but negative_prompt (allow-empty)
	required := map[string]bool{}
	if arr, ok := schema["required"].([]any); ok {
		for _, name := range arr {
			required[fmt.Sprint(name)] = true
		}
	} else {
		t.Fatalf("required = %v", schema["required"])
	}
	if required["negative_prompt"] {
		t.Error("negative_prompt must not be required (allow-empty)")
	}
	for _, name := range []string{"positive_prompt", "seed", "steps", "cfg", "denoise", "width", "height", "unet_name", "lora_name_1", "lora_strength_1"} {
		if !required[name] {
			t.Errorf("%s should be required", name)
		}
	}

	// image parameter on the inpaint tool
	inp, ok := byName["flux2_klein_9b_inpaint"]
	if !ok {
		t.Fatalf("inpaint tool missing: %v", byName)
	}
	inpSchema, _ := inp.InputSchema.(map[string]any)
	inpProps, _ := inpSchema["properties"].(map[string]any)
	if p, _ := inpProps["mask"].(map[string]any); p["type"] != "string" {
		t.Errorf("mask image param schema = %v", p)
	}
	if r0, _ := inpSchema["required"].([]any); len(r0) != 4 { // positive_prompt, crop_width, crop_height, mask
		t.Errorf("inpaint required count = %d (%v)", len(r0), r0)
	}
}

// --- tool calls ---

func TestCallTool_Txt2Img_FillsAndReturnsOutputs(t *testing.T) {
	r := newRig(t, map[string][]byte{
		"ComfyUI_00001_.png": []byte("png-bytes"),
		"meta_00001_.txt":    []byte("hello"),
	})

	res := callTool(t, r.session, "z_image_turbo_txt2img", validZitArgs())
	if res.IsError {
		t.Fatalf("tool error: %s", textOf(t, res))
	}

	// submitted workflow carries the values, no raw tokens
	submitted := r.submitted()
	raw, _ := json.Marshal(submitted["workflow"])
	s := string(raw)
	if strings.Contains(s, "{{") {
		t.Fatalf("raw token submitted: %s", s)
	}
	for _, want := range []string{`"seed":7`, `"steps":16`, `"width":512`, `"strength_model":0.8`, `"filename_prefix":"prefix-7"`} {
		if !strings.Contains(s, want) {
			t.Errorf("submitted workflow missing %s\n%s", want, s)
		}
	}
	// allow-empty token omitted from args lands as ""
	if !strings.Contains(s, `"text":""`) {
		t.Errorf("negative_prompt default not applied: %s", s)
	}
	// no images for this workflow
	if inputs, _ := submitted["inputs"].(map[string]any); len(inputs) != 0 {
		t.Errorf("unexpected inputs: %v", inputs)
	}

	// outputs written to disk
	txt := textOf(t, res)
	for _, name := range []string{"ComfyUI_00001_.png", "meta_00001_.txt"} {
		if !strings.Contains(txt, name) {
			t.Errorf("summary missing %s: %s", name, txt)
		}
	}
	dirEntries, _ := os.ReadDir(r.outputsDir)
	if len(dirEntries) == 0 {
		t.Fatal("no run dir created")
	}
	runDir := filepath.Join(r.outputsDir, dirEntries[0].Name())
	files, _ := os.ReadDir(runDir)
	if len(files) != 2 {
		t.Fatalf("expected 2 output files, got %d", len(files))
	}

	// summary + inlined png (the txt output is file-only)
	if len(res.Content) != 2 {
		t.Fatalf("content = %d items, want 2 (text + png)", len(res.Content))
	}
	img, ok := res.Content[1].(*mcp.ImageContent)
	if !ok {
		t.Fatalf("content[1] = %T, want ImageContent", res.Content[1])
	}
	if img.MIMEType != "image/png" || string(img.Data) != "png-bytes" {
		t.Errorf("image content wrong: mime=%s len=%d", img.MIMEType, len(img.Data))
	}
}

func TestCallTool_WithImageUploadsBase64(t *testing.T) {
	r := newRig(t, map[string][]byte{"out.png": []byte("x")})

	png := []byte("fake-png-data")
	pngPath := filepath.Join(t.TempDir(), "local.png")
	if err := os.WriteFile(pngPath, png, 0o644); err != nil {
		t.Fatal(err)
	}

	res := callTool(t, r.session, "flux2_klein_9b_inpaint", map[string]any{
		"positive_prompt": "fill the mask",
		"crop_width":      512,
		"crop_height":     512,
		"mask":            pngPath,
	})
	if res.IsError {
		t.Fatalf("tool error: %s", textOf(t, res))
	}

	submitted := r.submitted()
	inputs, _ := submitted["inputs"].(map[string]any)
	if len(inputs) != 1 {
		t.Fatalf("expected 1 uploaded input, got %v", inputs)
	}
	gotB64, _ := inputs["mask.png"].(string)
	want := base64.StdEncoding.EncodeToString(png)
	if gotB64 != want {
		t.Errorf("uploaded bytes mismatch: got %d b64 chars, want %d", len(gotB64), len(want))
	}
}

func TestCallTool_Errors(t *testing.T) {
	r := newRig(t, map[string][]byte{"out.png": []byte("x")})

	t.Run("missing token", func(t *testing.T) {
		args := validZitArgs()
		delete(args, "steps")
		res := callTool(t, r.session, "z_image_turbo_txt2img", args)
		if !res.IsError || !strings.Contains(textOf(t, res), "steps") {
			t.Fatalf("expected missing-steps error, got: %s", textOf(t, res))
		}
	})

	t.Run("non-numeric value", func(t *testing.T) {
		args := validZitArgs()
		args["steps"] = "walls"
		res := callTool(t, r.session, "z_image_turbo_txt2img", args)
		if !res.IsError || !strings.Contains(textOf(t, res), "not a number") {
			t.Fatalf("expected not-a-number error, got: %s", textOf(t, res))
		}
	})

	t.Run("numeric string is coerced", func(t *testing.T) {
		args := validZitArgs()
		args["steps"] = "16"
		res := callTool(t, r.session, "z_image_turbo_txt2img", args)
		if res.IsError {
			t.Fatalf("unexpected error: %s", textOf(t, res))
		}
		raw, _ := json.Marshal(r.submitted()["workflow"])
		if !strings.Contains(string(raw), `"steps":16`) {
			t.Errorf("string step not coerced: %s", raw)
		}
	})

	t.Run("missing image", func(t *testing.T) {
		res := callTool(t, r.session, "flux2_klein_9b_inpaint", map[string]any{
			"positive_prompt": "x", "crop_width": 512, "crop_height": 512,
		})
		if !res.IsError || !strings.Contains(textOf(t, res), "missing input image") {
			t.Fatalf("expected missing-image error, got: %s", textOf(t, res))
		}
	})

	t.Run("unreadable image path", func(t *testing.T) {
		res := callTool(t, r.session, "flux2_klein_9b_inpaint", map[string]any{
			"positive_prompt": "x", "crop_width": 512, "crop_height": 512,
			"mask": "/does/not/exist.png",
		})
		if !res.IsError || !strings.Contains(textOf(t, res), "read input image") {
			t.Fatalf("expected read-image error, got: %s", textOf(t, res))
		}
	})

	t.Run("unknown tool", func(t *testing.T) {
		_, err := r.session.CallTool(context.Background(), &mcp.CallToolParams{Name: "nope", Arguments: map[string]any{}})
		if err == nil {
			t.Fatal("expected error for unknown tool")
		}
	})
}

func TestCallTool_HeadlessErrorPropagates(t *testing.T) {
	// dedicated rig with a submit endpoint that 500s
	storeDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(storeDir, "z-image-turbo-txt2img.workflow.json"), []byte(zitFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	workflows, err := store.Load(storeDir)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/infer-async", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := &headless.Client{
		HTTP: srv.Client(), SubmitURL: srv.URL + "/infer-async", StatusURL: srv.URL + "/status",
		WorkflowsDir: storeDir, PollInterval: 5 * time.Millisecond, RunTimeout: 3 * time.Second,
	}
	server := Build(client, workflows, t.TempDir())

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	go server.Run(ctx, serverTransport)

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "1"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	res := callTool(t, session, "z_image_turbo_txt2img", validZitArgs())
	if !res.IsError || !strings.Contains(textOf(t, res), "HTTP 500") {
		t.Fatalf("expected HTTP 500 propagation, got: %s", textOf(t, res))
	}
}
