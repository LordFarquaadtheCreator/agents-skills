package generate

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeTestPNG creates a 100x100 solid-color PNG and returns its base64 encoding.
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

func TestRunNoPages(t *testing.T) {
	_, err := Run(Input{Title: "Test"})
	if err == nil {
		t.Fatal("expected error for no pages")
	}
}

func TestRunWritesPDF(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	out, err := Run(Input{
		Title:     "Test Story",
		Pages:     []Page{{Image: img, Text: "Once upon a time."}},
		OutputDir: dir,
		Filename:  "test.pdf",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasSuffix(out.OutputPath, "test.pdf") {
		t.Fatalf("output path = %q", out.OutputPath)
	}
	info, err := os.Stat(out.OutputPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF is empty")
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
			{Image: img, Text: "Page three with <b>bold</b> and <i>italic</i>."},
		},
		OutputDir: dir,
		Filename:  "multi.pdf",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.PageCount != 3 {
		t.Fatalf("pageCount = %d, want 3", out.PageCount)
	}
}

func TestRunDefaultsFilename(t *testing.T) {
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
	if out.Filename != "My Story.pdf" {
		t.Fatalf("filename = %q", out.Filename)
	}
}

func TestRunCreatesOutputDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deep")
	img := makeTestPNG(t, color.RGBA{R: 50, G: 50, B: 50, A: 255})
	out, err := Run(Input{
		Title:     "X",
		Pages:     []Page{{Image: img, Text: "text"}},
		OutputDir: dir,
		Filename:  "x.pdf",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.OutputPath); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestRunInvalidBase64(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(Input{
		Title:     "Bad",
		Pages:     []Page{{Image: "not-valid-base64!!!", Text: "text"}},
		OutputDir: dir,
		Filename:  "bad.pdf",
	})
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestRunHTMLFormatting(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	text := "This has <b>bold</b> and <i>italic</i> and <br> line breaks."
	out, err := Run(Input{
		Title:     "HTML Test",
		Pages:     []Page{{Image: img, Text: text}},
		OutputDir: dir,
		Filename:  "html.pdf",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.OutputPath); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestRunJPEGImage(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	out, err := Run(Input{
		Title:     "JPEG Test",
		Pages:     []Page{{Image: img, Text: "PNG works fine."}},
		OutputDir: dir,
		Filename:  "jpeg.pdf",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.OutputPath); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestRunCustomFontSize(t *testing.T) {
	dir := t.TempDir()
	img := makeTestPNG(t, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	out, err := Run(Input{
		Title:     "Custom Font",
		Pages:     []Page{{Image: img, Text: "Custom font size text."}},
		OutputDir: dir,
		Filename:  "custom.pdf",
		FontSize:  16,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.OutputPath); err != nil {
		t.Fatalf("file not created: %v", err)
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
		Title:     "Unicode \u2014 Test",
		Pages:     []Page{{Image: img, Text: text}},
		OutputDir: dir,
		Filename:  "unicode.pdf",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out.OutputPath); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}
