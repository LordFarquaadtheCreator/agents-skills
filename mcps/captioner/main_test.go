package main

import (
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/opentype"
)

func makeTestImage(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / w),
				G: uint8(y * 255 / h),
				B: 128,
				A: 255,
			})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestConstructCaptionedImage(t *testing.T) {
	tmpDir := t.TempDir()

	fontBytes, err := os.ReadFile("/Library/Fonts/Arial Unicode.ttf")
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	f, err := opentype.Parse(fontBytes)
	if err != nil {
		t.Fatalf("parse font: %v", err)
	}

	srcPath := filepath.Join(tmpDir, "test_input.png")
	makeTestImage(t, srcPath, 720, 480)

	cases := []struct {
		name string
		text string
	}{
		{"short", "Hello world."},
		{"medium", "A sleek white device sits on the table between them. \"Wear it eight hours a day,\" she says. \"It learns your stress patterns and optimizes everything.\""},
		{"long", "He signs the tablet without reading the fine print. A toggle at the bottom reads \"Reveal full optimization parameters\" — it's unchecked. He taps accept. The device on the table glows green for the first time. He doesn't know what he's just agreed to, and the room feels different now, charged with something he can't name."},
	}

	outDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img, err := constructCaptionedImage(srcPath, tc.text, f)
			if err != nil {
				t.Fatalf("constructCaptionedImage: %v", err)
			}

			bounds := img.Bounds()
			if bounds.Dy() <= 480 {
				t.Errorf("expected height > 480 (original), got %d", bounds.Dy())
			}

			outPath := filepath.Join(outDir, tc.name+".png")
			out, err := os.Create(outPath)
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			if err := png.Encode(out, img); err != nil {
				t.Fatal(err)
			}

			inspectDir := "/tmp/caption_test_inspect"
			os.MkdirAll(inspectDir, 0755)
			inspectPath := filepath.Join(inspectDir, tc.name+".png")
			inspectOut, err := os.Create(inspectPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(inspectOut, img); err != nil {
				inspectOut.Close()
				t.Fatal(err)
			}
			inspectOut.Close()

			t.Logf("output: %s (%dx%d)", outPath, bounds.Dx(), bounds.Dy())
			t.Logf("inspect: %s", inspectPath)
		})
	}
}

func TestConstructCaptionedImageVertical(t *testing.T) {
	tmpDir := t.TempDir()

	fontBytes, err := os.ReadFile("/Library/Fonts/Arial Unicode.ttf")
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	f, err := opentype.Parse(fontBytes)
	if err != nil {
		t.Fatalf("parse font: %v", err)
	}

	srcPath := filepath.Join(tmpDir, "vertical_input.png")
	makeTestImage(t, srcPath, 480, 720)

	words := []string{"device", "glows", "green", "compliance", "protocol", "optimize", "transition", "mirror", "nightstand", "pulses", "soft", "blue", "adjusting", "cortisol", "levels", "detected", "elevated", "sleep", "latency", "ambient"}

	rng := rand.New(rand.NewSource(42))
	textLen := 500 + rng.Intn(300)
	var sb []string
	total := 0
	for total < textLen {
		w := words[rng.Intn(len(words))]
		sb = append(sb, w)
		total += len(w) + 1
	}
	caption := ""
	for i, w := range sb {
		if i > 0 {
			caption += " "
		}
		caption += w
	}

	t.Logf("generated caption (%d chars): %s", len(caption), caption)

	img, err := constructCaptionedImage(srcPath, caption, f)
	if err != nil {
		t.Fatalf("constructCaptionedImage: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dy() <= 720 {
		t.Errorf("expected height > 720 (original), got %d", bounds.Dy())
	}
	if bounds.Dx() != 480 {
		t.Errorf("expected width 480, got %d", bounds.Dx())
	}

	outPath := filepath.Join(tmpDir, "vertical_output.png")
	out, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		t.Fatal(err)
	}

	inspectDir := "/tmp/caption_test_inspect"
	os.MkdirAll(inspectDir, 0755)
	inspectPath := filepath.Join(inspectDir, "vertical.png")
	inspectOut, err := os.Create(inspectPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(inspectOut, img); err != nil {
		inspectOut.Close()
		t.Fatal(err)
	}
	inspectOut.Close()

	t.Logf("output: %s (%dx%d)", outPath, bounds.Dx(), bounds.Dy())
	t.Logf("inspect: %s", inspectPath)
}

func TestConstructCaptionedImageVertical1000(t *testing.T) {
	tmpDir := t.TempDir()

	fontBytes, err := os.ReadFile("/Library/Fonts/Arial Unicode.ttf")
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	f, err := opentype.Parse(fontBytes)
	if err != nil {
		t.Fatalf("parse font: %v", err)
	}

	srcPath := filepath.Join(tmpDir, "vertical_1000_input.png")
	makeTestImage(t, srcPath, 480, 720)

	words := []string{"device", "glows", "green", "compliance", "protocol", "optimize", "transition", "mirror", "nightstand", "pulses", "soft", "blue", "adjusting", "cortisol", "levels", "detected", "elevated", "sleep", "latency", "ambient"}

	rng := rand.New(rand.NewSource(99))
	var sb []string
	total := 0
	for total < 1000 {
		w := words[rng.Intn(len(words))]
		sb = append(sb, w)
		total += len(w) + 1
	}
	caption := ""
	for i, w := range sb {
		if i > 0 {
			caption += " "
		}
		caption += w
	}

	t.Logf("generated caption (%d chars): %s", len(caption), caption)

	img, err := constructCaptionedImage(srcPath, caption, f)
	if err != nil {
		t.Fatalf("constructCaptionedImage: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dy() <= 720 {
		t.Errorf("expected height > 720 (original), got %d", bounds.Dy())
	}
	if bounds.Dx() != 480 {
		t.Errorf("expected width 480, got %d", bounds.Dx())
	}

	outPath := filepath.Join(tmpDir, "vertical_1000_output.png")
	out, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		t.Fatal(err)
	}

	inspectDir := "/tmp/caption_test_inspect"
	os.MkdirAll(inspectDir, 0755)
	inspectPath := filepath.Join(inspectDir, "vertical_1000.png")
	inspectOut, err := os.Create(inspectPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(inspectOut, img); err != nil {
		inspectOut.Close()
		t.Fatal(err)
	}
	inspectOut.Close()

	t.Logf("output: %s (%dx%d)", outPath, bounds.Dx(), bounds.Dy())
	t.Logf("inspect: %s", inspectPath)
}
