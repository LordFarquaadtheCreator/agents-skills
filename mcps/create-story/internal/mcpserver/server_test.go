package mcpserver

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/LordFarquaadtheCreator/create-story/internal/generate"
)

func makeTestPNG(t *testing.T, c color.Color) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func TestHandleGenerate(t *testing.T) {
	img := makeTestPNG(t, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	_, out, err := handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{
		Title:    "Test Story",
		Pages:    []generate.Page{{Image: img, Text: "Once upon a time."}},
		OutputDir: t.TempDir(),
		Filename:  "test.pdf",
	})
	if err != nil {
		t.Fatalf("handleGenerate: %v", err)
	}
	if !strings.HasSuffix(out.OutputPath, ".pdf") {
		t.Fatalf("output path = %q", out.OutputPath)
	}
	if out.PageCount != 1 {
		t.Fatalf("pageCount = %d, want 1", out.PageCount)
	}
}

func TestHandleGenerateMissingTitle(t *testing.T) {
	img := makeTestPNG(t, color.RGBA{R: 50, G: 50, B: 50, A: 255})
	_, _, err := handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{
		Pages: []generate.Page{{Image: img, Text: "text"}},
	})
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestHandleGenerateNoPages(t *testing.T) {
	_, _, err := handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{
		Title: "Test",
	})
	if err == nil {
		t.Fatal("expected error for no pages")
	}
}

func TestHandleGenerateMissingImage(t *testing.T) {
	_, _, err := handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{
		Title: "Test",
		Pages: []generate.Page{{Text: "text"}},
	})
	if err == nil {
		t.Fatal("expected error for missing image")
	}
}

func TestHandleGenerateMissingText(t *testing.T) {
	img := makeTestPNG(t, color.RGBA{R: 50, G: 50, B: 50, A: 255})
	_, _, err := handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{
		Title: "Test",
		Pages: []generate.Page{{Image: img}},
	})
	if err == nil {
		t.Fatal("expected error for missing text")
	}
}

func TestHandleGenerateMultiplePages(t *testing.T) {
	img := makeTestPNG(t, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	_, out, err := handleGenerate(context.Background(), &mcp.CallToolRequest{}, generate.Input{
		Title: "Multi",
		Pages: []generate.Page{
			{Image: img, Text: "Page one."},
			{Image: img, Text: "Page two."},
		},
		OutputDir: t.TempDir(),
		Filename:  "multi.pdf",
	})
	if err != nil {
		t.Fatalf("handleGenerate: %v", err)
	}
	if out.PageCount != 2 {
		t.Fatalf("pageCount = %d, want 2", out.PageCount)
	}
}
