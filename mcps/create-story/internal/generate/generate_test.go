package generate

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeTestPNG creates a 100x100 solid-color PNG temp file and returns its path.
func makeTestPNG(t *testing.T, c color.Color) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.CreateTemp(t.TempDir(), "test-*.png")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestRunNoPages(t *testing.T) {
	_, err := Run(Input{Title: "Test"})
	if err == nil {
		t.Fatal("expected error for no pages")
	}
}

func TestRunWritesPDFAndPNGs(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	out, err := Run(Input{
		Title:     "Test Story",
		Pages:     []Page{{Image: img, Text: "Once upon a time."}},
		OutputDir: dir,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasSuffix(out.PDFPath, "Test Story.pdf") {
		t.Fatalf("pdfPath = %q", out.PDFPath)
	}
	if info, err := os.Stat(out.PDFPath); err != nil || info.Size() == 0 {
		t.Fatalf("PDF missing or empty: %v", err)
	}
	if len(out.PNGPaths) != 1 {
		t.Fatalf("pngPaths len = %d, want 1", len(out.PNGPaths))
	}
	if info, err := os.Stat(out.PNGPaths[0]); err != nil || info.Size() == 0 {
		t.Fatalf("PNG missing or empty: %v", err)
	}
	if out.PageCount != 1 {
		t.Fatalf("pageCount = %d, want 1", out.PageCount)
	}
}

func TestRunMultiplePages(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	out, err := Run(Input{
		Title: "Multi",
		Pages: []Page{
			{Image: img, Text: "Page one text."},
			{Image: img, Text: "Page two text."},
			{Image: img, Text: "Page three with **bold** and *italic*."},
		},
		OutputDir: dir,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.PageCount != 3 {
		t.Fatalf("pageCount = %d, want 3", out.PageCount)
	}
	if len(out.PNGPaths) != 3 {
		t.Fatalf("pngPaths len = %d, want 3", len(out.PNGPaths))
	}
	for i, p := range out.PNGPaths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("png %d missing: %v", i+1, err)
		}
	}
}

func TestRunCreatesTitleSubdir(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 50, G: 50, B: 50, A: 255})
	out, err := Run(Input{
		Title:     "My Story",
		Pages:     []Page{{Image: img, Text: "text"}},
		OutputDir: dir,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasSuffix(out.OutputDir, "My Story") {
		t.Fatalf("outputDir = %q", out.OutputDir)
	}
}

func TestRunCollisionHandling(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 50, G: 50, B: 50, A: 255})

	out1, err := Run(Input{
		Title:     "Dup",
		Pages:     []Page{{Image: img, Text: "text"}},
		OutputDir: dir,
	})
	if err != nil {
		t.Fatalf("Run1: %v", err)
	}
	out2, err := Run(Input{
		Title:     "Dup",
		Pages:     []Page{{Image: img, Text: "text"}},
		OutputDir: dir,
	})
	if err != nil {
		t.Fatalf("Run2: %v", err)
	}
	if out1.OutputDir == out2.OutputDir {
		t.Fatalf("collision not handled: both got %q", out1.OutputDir)
	}
}

func TestRunInvalidImagePath(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(Input{
		Title:     "Bad",
		Pages:     []Page{{Image: "/nonexistent/image.png", Text: "text"}},
		OutputDir: dir,
	})
	if err == nil {
		t.Fatal("expected error for invalid image path")
	}
}

func TestRunMarkdownFormatting(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	text := "This has **bold** and *italic* and\nline breaks."
	out, err := Run(Input{
		Title:     "Markdown Test",
		Pages:     []Page{{Image: img, Text: text}},
		OutputDir: dir,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.PDFPath); err != nil {
		t.Fatalf("PDF not created: %v", err)
	}
}

func TestRunCustomFontSize(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	out, err := Run(Input{
		Title:     "Custom Font",
		Pages:     []Page{{Image: img, Text: "Custom font size text."}},
		OutputDir: dir,
		FontSize:  16,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.PDFPath); err != nil {
		t.Fatalf("PDF not created: %v", err)
	}
}

func TestSanitizeASCII(t *testing.T) {
	tests := []struct{ in, want string }{
		{"hello — world", "hello - world"},
		{"en–dash", "en-dash"},
		{"\u201Cquoted\u201D", `"quoted"`},
		{"\u2018single\u2019", "'single'"},
		{"ellipsis\u2026", "ellipsis..."},
		{"plain ascii", "plain ascii"},
		{"café — naïve", "caf - nave"},
	}
	for _, tc := range tests {
		got := sanitizeASCII(tc.in)
		if got != tc.want {
			t.Errorf("sanitizeASCII(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRunSanitizesUnicode(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	text := "He said \u201Chello\u201D \u2014 the naïve café\u2026"
	out, err := Run(Input{
		Title:     "Unicode Test",
		Pages:     []Page{{Image: img, Text: text}},
		OutputDir: dir,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.PDFPath); err != nil {
		t.Fatalf("PDF not created: %v", err)
	}
}

func TestRunNestedOutputDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deep")
	img := makeTestPNG(t, color.RGBA{R: 50, G: 50, B: 50, A: 255})
	out, err := Run(Input{
		Title:     "X",
		Pages:     []Page{{Image: img, Text: "text"}},
		OutputDir: dir,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.PDFPath); err != nil {
		t.Fatalf("PDF not created: %v", err)
	}
}

func TestRunPreviewPDF(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	out, err := Run(Input{
		Title: "Preview Test",
		Pages: []Page{
			{Image: img, Text: "Page one."},
			{Image: img, Text: "Page two."},
			{Image: img, Text: "Page three."},
		},
		OutputDir:        dir,
		PreviewAfterPage: 1,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.PreviewPDFPath == "" {
		t.Fatal("previewPDFPath is empty")
	}
	if _, err := os.Stat(out.PreviewPDFPath); err != nil {
		t.Fatalf("preview PDF not created: %v", err)
	}
	if !strings.HasSuffix(out.PreviewPDFPath, "_preview.pdf") {
		t.Fatalf("previewPDFPath = %q, want _preview.pdf suffix", out.PreviewPDFPath)
	}
	if len(out.PreviewPNGPaths) != 3 {
		t.Fatalf("previewPngPaths len = %d, want 3", len(out.PreviewPNGPaths))
	}
	for i, p := range out.PreviewPNGPaths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("preview png %d missing: %v", i+1, err)
		}
		if !strings.Contains(filepath.Base(p), "_preview_") {
			t.Fatalf("preview png %d path = %q, want _preview_ in name", i+1, p)
		}
	}
}

func TestRunNoPreviewByDefault(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	out, err := Run(Input{
		Title:     "No Preview",
		Pages:     []Page{{Image: img, Text: "text"}},
		OutputDir: dir,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.PreviewPDFPath != "" {
		t.Fatalf("previewPDFPath = %q, want empty", out.PreviewPDFPath)
	}
	if len(out.PreviewPNGPaths) != 0 {
		t.Fatalf("previewPngPaths len = %d, want 0", len(out.PreviewPNGPaths))
	}
}
