package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/LordFarquaadtheCreator/create-story/internal/generate"
)

// Run starts the stdio MCP server.
func Run() error {
	server := mcp.NewServer(&mcp.Implementation{Name: "create-story", Version: "1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_story_pdf",
		Description: "Generate a PDF book and per-page PNG images from image files and text. Each page has an image on the left and story text on the right, with a muted background color extracted from the image. The agent provides absolute file paths to PNG or JPEG images and text per page. Text supports markdown: **bold**, *italic*, \\n for line breaks, \\n\\n for paragraph breaks. Output goes to ~/Desktop/<title>/ — contains <title>.pdf and <title>.<n>.png for each page. Returns output directory, PDF path, and PNG paths.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args generate.Input) (*mcp.CallToolResult, generate.Output, error) {
		return handleGenerate(ctx, req, args)
	})

	return server.Run(context.Background(), &mcp.StdioTransport{})
}

func handleGenerate(ctx context.Context, req *mcp.CallToolRequest, args generate.Input) (*mcp.CallToolResult, generate.Output, error) {
	if args.Title == "" {
		return nil, generate.Output{}, fmt.Errorf("title is required")
	}
	if len(args.Pages) == 0 {
		return nil, generate.Output{}, fmt.Errorf("at least one page is required — each page needs an image file path and text")
	}
	for i, p := range args.Pages {
		if p.Image == "" {
			return nil, generate.Output{}, fmt.Errorf("page %d: image is required (absolute file path)", i+1)
		}
		if p.Text == "" {
			return nil, generate.Output{}, fmt.Errorf("page %d: text is required", i+1)
		}
	}

	out, err := generate.Run(args)
	if err != nil {
		return nil, generate.Output{}, err
	}
	return jsonResult(out)
}

// jsonResult marshals the structured output as pretty JSON in the text content.
func jsonResult[T any](out T) (*mcp.CallToolResult, T, error) {
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, out, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}, out, nil
}
