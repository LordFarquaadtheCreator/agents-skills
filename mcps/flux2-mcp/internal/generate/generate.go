package generate

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mcps/flux2-mcp/internal/fluxapi"
)

const DefaultOutputDir = "./output.private/mcp_output"

type Input struct {
	Prompt              string     `json:"prompt" jsonschema:"required,Text prompt for generation. Be descriptive — include style, subject, composition, lighting."`
	ReferenceImagePaths []string   `json:"reference_image_paths,omitempty" jsonschema:"Local file paths to reference images (PNG/JPG/WebP/AVIF). Server reads and encodes them. 2-3 recommended for best quality. More refs = slower (quadratic attention)."`
	Loras               []LoraSpec `json:"loras,omitempty" jsonschema:"List of LoRA specs: [{\"path\":\"file.safetensors\",\"strength\":0.8}]. Files resolved against /mnt/ComfyUI/models/loras/ on the volume."`
	Variant             string     `json:"variant,omitempty" jsonschema:"Model variant: klein-9b (base, CFG works, 25+ steps), klein-9b-kv (KV-distilled, 4-step optimized, guidance baked in), klein-4b (smaller, distilled). Default: klein-9b."`
	Width               *int       `json:"width,omitempty" jsonschema:"Output width in pixels. Must be divisible by 16. Default: 1024."`
	Height              *int       `json:"height,omitempty" jsonschema:"Output height in pixels. Must be divisible by 16. Default: 1024."`
	NumInferenceSteps   *int       `json:"num_inference_steps,omitempty" jsonschema:"Denoising steps. Default: 25. KV model optimized for 4, base model benefits from 25+."`
	GuidanceScale       *float64   `json:"guidance_scale,omitempty" jsonschema:"CFG scale. 1.0 = no CFG (single forward pass, ~2x faster). Default: 1.0. Only effective on non-distilled variants (klein-9b)."`
	Seed                *int64     `json:"seed,omitempty" jsonschema:"Fixed seed for reproducibility. Omit for random."`
	NegativePrompt      string     `json:"negative_prompt,omitempty" jsonschema:"Negative prompt. Ignored when guidance_scale=1.0 (CFG disabled)."`
	OutputFilename      string     `json:"output_filename,omitempty" jsonschema:"Custom output filename (without extension). .png is forced. Default: auto-generated."`
	OutputMode          string     `json:"output_mode,omitempty" jsonschema:"How to return images: file (save to disk), base64 (return inline), both. Default: file."`
	OutputDir           string     `json:"output_dir,omitempty" jsonschema:"Directory for saved images. Default: ./output.private/mcp_output"`
	Repeat              int        `json:"repeat,omitempty" jsonschema:"Generate N images with incrementing seeds. Default: 1."`
}

type LoraSpec = fluxapi.LoraSpec

type ImageResult struct {
	Path   string `json:"path,omitempty"`
	Seed   string `json:"seed"`
	Bytes  int64  `json:"bytes"`
	Inline bool   `json:"inline"`
}

type Output struct {
	Message  string        `json:"message"`
	Images   []ImageResult `json:"images"`
	Warnings []string      `json:"warnings,omitempty"`
}

type Result struct {
	Path  string
	Seed  string
	Bytes int64
	Data  []byte
}

type APIClient interface {
	Generate(ctx context.Context, req fluxapi.GenerateRequest) ([]byte, error)
}

var validOutputModes = map[string]bool{"file": true, "base64": true, "both": true, "": true}

const seedStep = 10

func Validate(args Input) error {
	if args.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if !validOutputModes[args.OutputMode] {
		return fmt.Errorf("invalid output_mode %q: must be file, base64, or both", args.OutputMode)
	}
	if args.Width != nil && *args.Width%16 != 0 {
		return fmt.Errorf("width must be divisible by 16, got %d", *args.Width)
	}
	if args.Height != nil && *args.Height%16 != 0 {
		return fmt.Errorf("height must be divisible by 16, got %d", *args.Height)
	}
	return nil
}

func Run(ctx context.Context, client APIClient, args Input) ([]Result, []string, error) {
	log.Printf("[flux2] starting generation: prompt=%q refs=%d loras=%d repeat=%d",
		truncate(args.Prompt, 80), len(args.ReferenceImagePaths), len(args.Loras), max(args.Repeat, 1))

	if err := Validate(args); err != nil {
		return nil, nil, err
	}

	// resolve reference image paths to base64
	var refImages []string
	if len(args.ReferenceImagePaths) > 0 {
		log.Printf("[flux2] loading %d reference images from disk", len(args.ReferenceImagePaths))
		for _, p := range args.ReferenceImagePaths {
			b64, err := EncodeImageFile(p)
			if err != nil {
				return nil, nil, fmt.Errorf("load ref image %s: %w", p, err)
			}
			refImages = append(refImages, b64)
			log.Printf("[flux2]   encoded %s -> %d base64 chars", p, len(b64))
		}
	}

	mode := args.OutputMode
	if mode == "" {
		mode = "file"
	}

	outputDir := args.OutputDir
	if outputDir == "" {
		outputDir = DefaultOutputDir
	}

	var warnings []string
	if args.GuidanceScale != nil && *args.GuidanceScale == 1.0 && args.NegativePrompt != "" {
		warnings = append(warnings, "negative_prompt has no effect when guidance_scale=1.0 (CFG disabled)")
	}

	repeat := max(args.Repeat, 1)
	results := make([]Result, 0, repeat)

	for i := range repeat {
		apiReq := buildAPIRequest(args, i, refImages)
		log.Printf("[flux2] iteration %d/%d: seed=%v width=%d height=%d steps=%d guidance=%.1f",
			i+1, repeat, seedPtr(apiReq.Seed), apiReq.Width, apiReq.Height, apiReq.NumInferenceSteps, apiReq.GuidanceScale)

		t0 := time.Now()
		data, err := client.Generate(ctx, apiReq)
		elapsed := time.Since(t0)
		if err != nil {
			log.Printf("[flux2] iteration %d failed after %v: %v", i+1, elapsed, err)
			return nil, nil, err
		}
		log.Printf("[flux2] iteration %d done in %v: %d bytes", i+1, elapsed, len(data))

		r := Result{Bytes: int64(len(data))}

		if mode == "file" || mode == "both" {
			filename := buildFilename(args.OutputFilename, i)
			outputPath := filepath.Join(outputDir, filename)
			if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
				return nil, nil, fmt.Errorf("create output dir: %w", err)
			}
			if err := os.WriteFile(outputPath, data, 0o644); err != nil {
				return nil, nil, fmt.Errorf("write output file: %w", err)
			}
			r.Path = outputPath
			log.Printf("[flux2] saved to %s", outputPath)
		}

		if mode == "base64" || mode == "both" {
			r.Data = data
		}

		if apiReq.Seed != nil {
			r.Seed = fmt.Sprintf("%d", *apiReq.Seed)
		} else {
			r.Seed = "random"
		}

		results = append(results, r)
	}

	return results, warnings, nil
}

func FormatResults(results []Result) string {
	lines := make([]string, len(results))
	for i, r := range results {
		if r.Path != "" {
			lines[i] = fmt.Sprintf("[%d/%d] %s (%d bytes) seed=%s", i+1, len(results), r.Path, r.Bytes, r.Seed)
		} else {
			lines[i] = fmt.Sprintf("[%d/%d] inline (%d bytes) seed=%s", i+1, len(results), r.Bytes, r.Seed)
		}
	}
	return fmt.Sprintf("Generated %d image(s):\n%s", len(results), strings.Join(lines, "\n"))
}

func ToOutput(results []Result, warnings []string) Output {
	images := make([]ImageResult, len(results))
	for i, r := range results {
		images[i] = ImageResult{
			Path:   r.Path,
			Seed:   r.Seed,
			Bytes:  r.Bytes,
			Inline: r.Data != nil,
		}
	}
	msg := FormatResults(results)
	if len(warnings) > 0 {
		msg += "\n\nWarnings:\n" + strings.Join(warnings, "\n")
	}
	return Output{Message: msg, Images: images, Warnings: warnings}
}

func buildAPIRequest(args Input, iteration int, refImages []string) fluxapi.GenerateRequest {
	w := 1024
	if args.Width != nil {
		w = *args.Width
	}
	h := 1024
	if args.Height != nil {
		h = *args.Height
	}
	steps := 25
	if args.NumInferenceSteps != nil {
		steps = *args.NumInferenceSteps
	}
	guidance := 1.0
	if args.GuidanceScale != nil {
		guidance = *args.GuidanceScale
	}

	req := fluxapi.GenerateRequest{
		Prompt:            args.Prompt,
		ReferenceImages:   refImages,
		Loras:             args.Loras,
		Width:             w,
		Height:            h,
		NumInferenceSteps: steps,
		GuidanceScale:     guidance,
		NegativePrompt:    args.NegativePrompt,
		Variant:           args.Variant,
	}

	if args.Seed != nil && *args.Seed > 0 {
		seed := *args.Seed + int64(iteration)*int64(seedStep)
		req.Seed = &seed
	}

	return req
}

func buildFilename(outputFilename string, iteration int) string {
	filename := outputFilename
	if filename == "" {
		filename = "flux2_output"
	}
	filename = stripExt(filename) + ".png"
	if iteration > 0 {
		filename = versionedFilename(filename, iteration+1)
	}
	return filename
}

func stripExt(name string) string {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext)
}

func versionedFilename(name string, version int) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s_v%d%s", base, version, ext)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func seedPtr(s *int64) string {
	if s == nil {
		return "random"
	}
	return fmt.Sprintf("%d", *s)
}

// EncodeImageFile reads a file and returns base64-encoded raw bytes.
// Python backend handles decode/resize for all formats (JPEG, PNG, WebP, AVIF, GIF).
func EncodeImageFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read image file %s: %w", path, err)
	}
	log.Printf("[flux2]   loaded %s (%d bytes)", path, len(data))
	return base64.StdEncoding.EncodeToString(data), nil
}
