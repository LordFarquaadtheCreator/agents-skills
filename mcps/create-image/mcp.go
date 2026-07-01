package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

const (
	defaultRequestTimeout = 300 * time.Second
	defaultSeedStep       = 10
	modelCardPath         = "model_card.yaml"
	defaultOutputDir      = "./output/mcp_output"
)

// Hardcoded defaults from model_card.yaml. These are not exposed as MCP params.
const (
	defaultCfg         = 1.0
	defaultSamplerName = "euler"
	defaultScheduler   = "simple"
	defaultDenoise     = 1.0
	defaultUnetName    = "z_image_turbo_bf16 (1).safetensors"
)

func runMCPServer(apiURL string) error {
	client := &http.Client{Timeout: defaultRequestTimeout}

	server := mcp.NewServer(&mcp.Implementation{Name: "create-image", Version: "1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_loras",
		Description: "List all available LoRAs and base models from model_card.yaml for the Modal ComfyUI generation workflow. Returns filename, name, prompt_style, keywords, notes, link, recommended_strength, etc.",
	}, handleListLoras)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_image",
		Description: "Generate an image using the Modal ComfyUI generation workflow. Returns the saved file path and metadata.",
	}, func(ctx context.Context, ss *mcp.ServerSession, req *mcp.CallToolParamsFor[GenerateImageInput]) (*mcp.CallToolResultFor[GenerateImageOutput], error) {
		return handleGenerateImage(ctx, client, apiURL, req.Arguments)
	})

	return server.Run(context.Background(), &mcp.StdioTransport{})
}

func handleListLoras(ctx context.Context, ss *mcp.ServerSession, req *mcp.CallToolParamsFor[struct{}]) (*mcp.CallToolResultFor[ListLorasOutput], error) {
	data, err := os.ReadFile(modelCardPath)
	if err != nil {
		return nil, fmt.Errorf("read model_card.yaml: %w", err)
	}

	var card ModelCard
	if err := yaml.Unmarshal(data, &card); err != nil {
		return nil, fmt.Errorf("parse model_card.yaml: %w", err)
	}

	b, err := json.MarshalIndent(card.Loras, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal loras: %w", err)
	}

	return &mcp.CallToolResultFor[ListLorasOutput]{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(b)}},
		StructuredContent: ListLorasOutput{Loras: card.Loras},
	}, nil
}

func handleGenerateImage(ctx context.Context, client *http.Client, apiURL string, args GenerateImageInput) (*mcp.CallToolResultFor[GenerateImageOutput], error) {
	if args.PositivePrompt == "" {
		return nil, fmt.Errorf("positive_prompt is required")
	}
	if args.LoraFilename1 == "" || args.LoraFilename2 == "" || args.LoraFilename3 == "" {
		return nil, fmt.Errorf("all three lora filenames are required")
	}

	repeat := args.Repeat
	if repeat < 1 {
		repeat = 1
	}

	seedStep := defaultSeedStep

	var lines []string

	for i := 0; i < repeat; i++ {
		apiReq := buildAPIRequest(args, i, seedStep)

		filename := buildFilename(args.OutputFilename, i)
		outputPath := filepath.Join(defaultOutputDir, filename)

		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return nil, fmt.Errorf("create output dir: %w", err)
		}

		n, err := postAndSave(ctx, client, apiURL, apiReq, outputPath)
		if err != nil {
			return nil, err
		}

		seedStr := "N/A"
		if apiReq.Seed != nil {
			seedStr = fmt.Sprintf("%d", *apiReq.Seed)
		}
		lines = append(lines, fmt.Sprintf("[%d/%d] %s (%d bytes) seed=%s", i+1, repeat, outputPath, n, seedStr))
	}

	msg := fmt.Sprintf("Generated %d image(s):\n%s", repeat, strings.Join(lines, "\n"))
	return &mcp.CallToolResultFor[GenerateImageOutput]{
		Content:           []mcp.Content{&mcp.TextContent{Text: msg}},
		StructuredContent: GenerateImageOutput{Message: msg},
	}, nil
}

func buildAPIRequest(args GenerateImageInput, iteration int, seedStep int) Request {
	req := Request{
		PositivePrompt: args.PositivePrompt,
		NegativePrompt: args.NegativePrompt,
		LoraName1:      args.LoraFilename1,
		LoraStrength1:  args.LoraStrength1,
		LoraName2:      args.LoraFilename2,
		LoraStrength2:  args.LoraStrength2,
		LoraName3:      args.LoraFilename3,
		LoraStrength3:  args.LoraStrength3,
	}

	// Hardcoded defaults (not exposed as MCP params)
	cfgVal := defaultCfg
	req.Cfg = &cfgVal
	sampler := defaultSamplerName
	req.SamplerName = &sampler
	sched := defaultScheduler
	req.Scheduler = &sched
	denoiseVal := defaultDenoise
	req.Denoise = &denoiseVal
	req.UnetName = defaultUnetName

	if args.Seed != nil && *args.Seed > 0 {
		seed := *args.Seed + int64(iteration)*int64(seedStep)
		req.Seed = &seed
	}
	if args.Steps != nil {
		req.Steps = args.Steps
	}
	if args.Width != nil {
		req.Width = args.Width
	}
	if args.Height != nil {
		req.Height = args.Height
	}

	return req
}

func buildFilename(outputFilename string, iteration int) string {
	filename := outputFilename
	if filename == "" {
		filename = "mcp_output"
	}
	// Strip any existing extension, then force .png
	filename = stripExt(filename) + ".png"
	if iteration > 0 {
		filename = versionedFilename(filename, iteration+1)
	}
	return filename
}

func postAndSave(ctx context.Context, client *http.Client, apiURL string, apiReq Request, outputPath string) (int64, error) {
	body, err := json.Marshal(apiReq)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := new(bytes.Buffer)
		_, _ = msg.ReadFrom(resp.Body)
		return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg.String())
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	n, err := out.ReadFrom(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("write output: %w", err)
	}
	return n, nil
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
