package mcpserver

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcps/flux2-mcp/internal/fluxapi"
	"mcps/flux2-mcp/internal/generate"
)

type deps struct {
	Client generate.APIClient
}

func newServer(d deps) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "flux2-mcp", Version: "1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_image",
		Description: "Generate an image using FLUX.2 Klein 9B with multi-reference image conditioning and LoRA support. Async generation — takes 25-45s on warm container, 60-90s cold. Use 2-3 reference images for best quality. guidance_scale=1.0 skips CFG (2x faster). LoRAs are merged into transformer weights on the fly.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args generate.Input) (*mcp.CallToolResult, generate.Output, error) {
		return handleGenerate(ctx, d.Client, args)
	})

	return server
}

func Run(apiURL string) error {
	return RunWithTransport(context.Background(), apiURL, &mcp.StdioTransport{})
}

func RunWithTransport(ctx context.Context, apiURL string, transport mcp.Transport) error {
	d := deps{Client: fluxapi.NewClient(apiURL)}
	server := newServer(d)
	return server.Run(ctx, transport)
}

func RunCLI(apiURL string, args []string) error {
	fs := flag.NewFlagSet("flux2-cli", flag.ContinueOnError)

	prompt := fs.String("prompt", "", "Text prompt for generation (required)")
	refDir := fs.String("ref-dir", "", "Directory of reference images to encode and send")
	maxRefs := fs.Int("max-refs", 0, "Limit reference images (0 = all)")
	variant := fs.String("variant", "klein-9b", "Model variant: klein-9b, klein-9b-kv, klein-4b")
	width := fs.Int("width", 1024, "Output width (divisible by 16)")
	height := fs.Int("height", 1024, "Output height (divisible by 16)")
	steps := fs.Int("steps", 25, "Denoising steps")
	guidance := fs.Float64("guidance", 1.0, "CFG scale (1.0 = no CFG)")
	seed := fs.Int64("seed", 0, "Random seed (0 = random)")
	negative := fs.String("negative", "", "Negative prompt (ignored when guidance=1.0)")
	output := fs.String("output", "output.png", "Output file path")
	outputMode := fs.String("output-mode", "file", "Output mode: file, base64, both")
	outputDir := fs.String("output-dir", "", "Output directory (default: same as output file)")
	repeat := fs.Int("repeat", 1, "Generate N images with incrementing seeds")
	loraPaths := fs.String("lora-paths", "", "Comma-separated LoRA filenames on volume (under models/loras/)")
	loraStrengths := fs.String("lora-strengths", "", "Comma-separated LoRA strengths (matches lora-paths order)")
	loraPath := fs.String("lora-path", "", "Single LoRA filename (shorthand for lora-paths)")
	loraStrength := fs.Float64("lora-strength", 1.0, "Single LoRA strength (used with lora-path)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *prompt == "" {
		return fmt.Errorf("--prompt is required")
	}

	input := generate.Input{
		Prompt:            *prompt,
		Variant:           *variant,
		Width:             width,
		Height:            height,
		NumInferenceSteps: steps,
		GuidanceScale:     guidance,
		NegativePrompt:    *negative,
		OutputMode:        *outputMode,
		Repeat:            *repeat,
	}

	if *outputDir != "" {
		input.OutputDir = *outputDir
	} else {
		input.OutputDir = filepath.Dir(*output)
	}
	input.OutputFilename = stripExt(filepath.Base(*output))

	if *seed > 0 {
		input.Seed = seed
	}

	if *refDir != "" {
		refs, err := loadReferenceImagePaths(*refDir, *maxRefs)
		if err != nil {
			return err
		}
		input.ReferenceImagePaths = refs
	}

	// Build LoRA list: prefer lora-paths, fall back to single lora-path
	if *loraPaths != "" {
		paths := strings.Split(*loraPaths, ",")
		var strengths []string
		if *loraStrengths != "" {
			strengths = strings.Split(*loraStrengths, ",")
		}
		for i, p := range paths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			s := 1.0
			if i < len(strengths) {
				if v, err := strconv.ParseFloat(strings.TrimSpace(strengths[i]), 64); err == nil {
					s = v
				}
			}
			input.Loras = append(input.Loras, generate.LoraSpec{Path: p, Strength: s})
		}
	} else if *loraPath != "" {
		input.Loras = []generate.LoraSpec{
			{Path: *loraPath, Strength: *loraStrength},
		}
	}

	client := fluxapi.NewClient(apiURL)
	results, warnings, err := generate.Run(context.Background(), client, input)
	if err != nil {
		return err
	}

	out := generate.ToOutput(results, warnings)
	fmt.Println(out.Message)
	return nil
}

func handleGenerate(ctx context.Context, client generate.APIClient, args generate.Input) (*mcp.CallToolResult, generate.Output, error) {
	results, warnings, err := generate.Run(ctx, client, args)
	if err != nil {
		return nil, generate.Output{}, err
	}

	out := generate.ToOutput(results, warnings)

	content := []mcp.Content{&mcp.TextContent{Text: out.Message}}
	for _, r := range results {
		if r.Data != nil {
			content = append(content, &mcp.ImageContent{
				Data:     r.Data,
				MIMEType: "image/png",
			})
		}
	}

	return &mcp.CallToolResult{Content: content}, out, nil
}

func loadReferenceImagePaths(dir string, maxRefs int) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read ref dir: %w", err)
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".avif" {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}

	if maxRefs > 0 && len(paths) > maxRefs {
		paths = paths[:maxRefs]
	}

	log.Printf("[flux2-cli] found %d reference images in %s", len(paths), dir)
	return paths, nil
}

func stripExt(name string) string {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext)
}
