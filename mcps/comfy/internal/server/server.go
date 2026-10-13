// Package server exposes the comfy-headless workflow store as MCP tools:
// one tool per workflow, with a generated input schema (one parameter per
// {{token}}, one local-file-path parameter per LoadImage input). Tool calls
// run synchronously against comfy-headless and return the produced files,
// with images inlined where the client supports it.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcps/comfy/internal/headless"
	"mcps/comfy/internal/store"
)

// Build assembles the MCP server from a loaded workflow set. Runs and tests
// share this; Run only adds the stdio transport.
func Build(client *headless.Client, workflows []store.Workflow, outputDir string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "comfy", Version: "1.0.0"}, nil)
	for _, w := range workflows {
		s.AddTool(toolFor(w), handlerFor(client, w, outputDir))
	}
	return s
}

// Run serves the tools over stdio.
func Run(client *headless.Client, workflows []store.Workflow, outputDir string) error {
	return Build(client, workflows, outputDir).Run(context.Background(), &mcp.StdioTransport{})
}

func toolFor(w store.Workflow) *mcp.Tool {
	return &mcp.Tool{
		Name:        w.Tool,
		Description: descriptionFor(w),
		InputSchema: schemaFor(w),
	}
}

func descriptionFor(w store.Workflow) string {
	return fmt.Sprintf(
		"Run the %s workflow on comfy-headless (store file %s). Parameters map 1:1 onto its {{token}} placeholders; image parameters take the path of a local file that is uploaded with the job. Renders take 30s-3min when warm; the first call after idle can cold-start for a few minutes.",
		w.Tool, w.File)
}

// schemaFor generates the JSON Schema for one workflow's tool: token
// parameters (number/string per the shared standard), image path parameters,
// and a required list. Tokens where an empty string is legal stay optional.
func schemaFor(w store.Workflow) map[string]any {
	props := map[string]any{}
	var required []string

	for _, t := range w.Tokens {
		desc := fmt.Sprintf("Value for the {{%s}} placeholder (%s).", t.Name, t.Type)
		if t.AllowEmpty {
			desc += " May be empty (\"\")."
		}
		props[t.Name] = map[string]any{"type": t.Type, "description": desc}
		if !t.AllowEmpty {
			required = append(required, t.Name)
		}
	}

	for _, img := range w.Images {
		props[img.Param] = map[string]any{
			"type":        "string",
			"description": fmt.Sprintf("Path to a local image file for the workflow's %q LoadImage input.", img.Filename),
		}
		required = append(required, img.Param)
	}

	return map[string]any{
		"type":                 "object",
		"properties":           props,
		"required":             required,
		"additionalProperties": false,
	}
}

func handlerFor(client *headless.Client, w store.Workflow, outputDir string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := execute(ctx, client, w, req, outputDir)
		if err != nil {
			// Expected failures (validation, missing images, headless errors)
			// are tool results, not protocol errors — every client renders
			// the text instead of dropping the session.
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil
		}
		return res, nil
	}
}

func execute(ctx context.Context, client *headless.Client, w store.Workflow, req *mcp.CallToolRequest, outputDir string) (*mcp.CallToolResult, error) {
	args, err := argsMap(req)
	if err != nil {
		return nil, err
	}

	values, err := collectValues(w, args)
	if err != nil {
		return nil, err
	}

	images, err := collectImages(w, args)
	if err != nil {
		return nil, err
	}

	outputs, err := client.Run(ctx, w.File, values, images)
	if err != nil {
		return nil, err
	}

	return resultFor(w, outputs, outputDir)
}

func argsMap(req *mcp.CallToolRequest) (map[string]any, error) {
	if req.Params == nil || len(req.Params.Arguments) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &m); err != nil {
		return nil, fmt.Errorf("parse arguments: %w", err)
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

// collectValues turns the call arguments into the workflow's value set.
// Missing values fail loudly unless the token allows empty (then "").
func collectValues(w store.Workflow, args map[string]any) (map[string]any, error) {
	values := make(map[string]any, len(w.Tokens))
	var missing []string
	for _, t := range w.Tokens {
		raw, ok := args[t.Name]
		if !ok || raw == nil {
			if t.AllowEmpty {
				values[t.Name] = ""
				continue
			}
			missing = append(missing, t.Name)
			continue
		}
		if t.Type == "number" {
			n, err := toNumber(raw)
			if err != nil {
				return nil, fmt.Errorf("token %q: %w", t.Name, err)
			}
			values[t.Name] = n
		} else {
			values[t.Name] = fmt.Sprint(raw)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required value(s): %s", strings.Join(missing, ", "))
	}
	return values, nil
}

func toNumber(raw any) (any, error) {
	switch n := raw.(type) {
	case float64:
		return n, nil
	case int:
		return n, nil
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return nil, fmt.Errorf("not a number: %q", n.String())
		}
		return f, nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return nil, fmt.Errorf("not a number: %q", n)
		}
		return f, nil
	}
	return nil, fmt.Errorf("not a number: %v", raw)
}

// collectImages reads each image parameter's file from disk.
func collectImages(w store.Workflow, args map[string]any) (map[string][]byte, error) {
	if len(w.Images) == 0 {
		return nil, nil
	}
	images := make(map[string][]byte, len(w.Images))
	for _, img := range w.Images {
		raw, ok := args[img.Param]
		path := ""
		if ok && raw != nil {
			path = strings.TrimSpace(fmt.Sprint(raw))
		}
		if path == "" {
			return nil, fmt.Errorf("missing input image: %s (pass a local file path for %q)", img.Param, img.Filename)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read input image %s: %w", path, err)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("input image %s is empty", path)
		}
		images[img.Filename] = data
	}
	return images, nil
}

// resultFor saves every output under outputDir and returns the MCP content:
// a summary text plus inline images.
func resultFor(w store.Workflow, outputs []headless.Output, outputDir string) (*mcp.CallToolResult, error) {
	dir := filepath.Join(outputDir, fmt.Sprintf("%s-%s", w.Tool, time.Now().Format("20060102-150405")))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	var summary strings.Builder
	fmt.Fprintf(&summary, "Ran %s: %d output file(s) saved to %s\n", w.File, len(outputs), dir)

	content := make([]mcp.Content, 0, len(outputs)+1)
	for _, o := range outputs {
		name := filepath.Base(o.Filename)
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, o.Data, 0o644); err != nil {
			return nil, fmt.Errorf("write output %s: %w", path, err)
		}
		fmt.Fprintf(&summary, "- %s (%d bytes) -> %s\n", name, len(o.Data), path)
		if mime := imageMime(name); mime != "" {
			content = append(content, &mcp.ImageContent{Data: o.Data, MIMEType: mime})
		}
	}

	return &mcp.CallToolResult{
		Content: append([]mcp.Content{&mcp.TextContent{Text: summary.String()}}, content...),
	}, nil
}

func imageMime(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	}
	return ""
}
