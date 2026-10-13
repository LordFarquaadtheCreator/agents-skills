// Package store loads the comfy-headless workflow store. Every
// *.workflow.json in the store directory becomes one MCP tool whose
// parameters are generated from the file itself: one parameter per
// {{token}} placeholder plus one local-file-path parameter per LoadImage
// input. Nothing here is hardcoded per workflow — drop a new file into the
// store and it becomes a tool on the next start.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var tokenRe = regexp.MustCompile(`\{\{(\w+)\}\}`)

// Numeric token names — the shared {{token}} standard's numeric set.
// Mirrors modal-ui/src/lib/run.ts (NUMERIC_TOKENS + NUMERIC_PATTERNS); the
// standard itself is documented in comfy-headless/workflows/README.md.
var numericTokens = map[string]bool{
	"seed": true, "steps": true, "cfg": true, "denoise": true,
	"width": true, "height": true, "length": true, "batch_size": true,
	"fps": true, "bit_depth": true, "shift": true, "megapixels": true,
	"pulid_strength": true, "face_index": true,
	"ref_width": true, "ref_height": true, "crop_width": true, "crop_height": true,
	"context_expand_factor": true, "blend_pixels": true,
}

var numericPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^strength_(model|clip)_\d+$`),
	regexp.MustCompile(`^lora_strength_\d+$`),
}

// Tokens where an empty string is a legal value (mirrors run.ts ALLOW_EMPTY).
var allowEmptyTokens = map[string]bool{
	"negative_prompt": true,
	"clothes":         true,
}

// TokenSpec is one workflow parameter derived from a {{token}}.
type TokenSpec struct {
	Name       string
	Type       string // "number" or "string"
	AllowEmpty bool
}

// ImageInput is one LoadImage input. Param is the MCP parameter name (a local
// file path); Filename is the upload key the workflow's LoadImage node reads.
type ImageInput struct {
	Param    string
	Filename string
}

// Workflow is one store file exposed as one MCP tool.
type Workflow struct {
	File   string // store filename, e.g. z-image-turbo-txt2img.workflow.json
	Tool   string // MCP tool name, e.g. z_image_turbo_txt2img
	Tokens []TokenSpec
	Images []ImageInput
}

// Load reads every *.workflow.json under dir. It fails loudly on anything
// ambiguous (name collisions, unparsable files, tokenless workflows) so a
// mis-shaped store never half-registers.
func Load(dir string) ([]Workflow, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read workflows dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".workflow.json") {
			files = append(files, e.Name())
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no *.workflow.json files in %s", dir)
	}
	sort.Strings(files)

	workflows := make([]Workflow, 0, len(files))
	seenTools := map[string]string{}
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f, err)
		}
		var template any
		if err := json.Unmarshal(raw, &template); err != nil {
			return nil, fmt.Errorf("parse %s: %w", f, err)
		}

		tool := ToolName(f)
		if prev, dup := seenTools[tool]; dup {
			return nil, fmt.Errorf("tool name collision: %s and %s both map to %q", prev, f, tool)
		}
		seenTools[tool] = f

		names := extractTokens(template)
		if len(names) == 0 {
			return nil, fmt.Errorf("%s has no {{token}} placeholders — refusing to expose a parameterless tool", f)
		}
		specs := make([]TokenSpec, 0, len(names))
		tokenNames := map[string]bool{}
		for _, name := range names {
			tokenNames[name] = true
			specs = append(specs, TokenSpec{Name: name, Type: tokenType(name), AllowEmpty: allowEmptyTokens[name]})
		}

		images := imageInputs(template, tokenNames)

		workflows = append(workflows, Workflow{File: f, Tool: tool, Tokens: specs, Images: images})
	}
	return workflows, nil
}

// ToolName maps a store filename to a tool name:
// "z-image-turbo-txt2img.workflow.json" -> "z_image_turbo_txt2img".
func ToolName(fileName string) string {
	base := strings.TrimSuffix(fileName, ".workflow.json")
	return slug(base)
}

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]+`)

func slug(s string) string {
	s = nonAlnum.ReplaceAllString(s, "_")
	return strings.Trim(s, "_")
}

func tokenType(name string) string {
	if numericTokens[name] {
		return "number"
	}
	for _, re := range numericPatterns {
		if re.MatchString(name) {
			return "number"
		}
	}
	return "string"
}

// extractTokens returns every token name in the template, sorted for
// determinism (Go map iteration is randomized).
func extractTokens(v any) []string {
	seen := map[string]bool{}
	var walk func(any)
	walk = func(v any) {
		switch t := v.(type) {
		case string:
			for _, m := range tokenRe.FindAllStringSubmatch(t, -1) {
				seen[m[1]] = true
			}
		case map[string]any:
			for _, vv := range t {
				walk(vv)
			}
		case []any:
			for _, vv := range t {
				walk(vv)
			}
		}
	}
	walk(v)

	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// imageInputs collects the distinct filenames referenced by LoadImage nodes,
// with a parameter name per filename. Two nodes reading the same file share
// one parameter. Parameters never collide with token names (suffix _image) or
// with each other (suffix _2, _3, ...).
func imageInputs(v any, tokenNames map[string]bool) []ImageInput {
	root, ok := v.(map[string]any)
	if !ok {
		return nil
	}

	seenFiles := map[string]bool{}
	var inputs []ImageInput
	for _, entry := range root {
		node, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if ct, _ := node["class_type"].(string); ct != "LoadImage" {
			continue
		}
		ins, _ := node["inputs"].(map[string]any)
		filename, _ := ins["image"].(string)
		if filename == "" || seenFiles[filename] {
			continue
		}
		seenFiles[filename] = true
		inputs = append(inputs, ImageInput{Param: "", Filename: filename})
	}
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].Filename < inputs[j].Filename })

	used := map[string]bool{}
	for i := range inputs {
		param := imageParamName(inputs[i].Filename)
		if tokenNames[param] {
			param += "_image"
		}
		for used[param] {
			param += "_2"
		}
		used[param] = true
		inputs[i].Param = param
	}
	return inputs
}

func imageParamName(filename string) string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	if slugged := slug(base); slugged != "" {
		return slugged
	}
	return "image"
}
