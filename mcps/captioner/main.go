package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type CaptionInput struct {
	Captions   []CaptionItem `json:"captions" jsonschema:"list of images with their caption text"`
	OutputPath string        `json:"outputPath,omitempty" jsonschema:"directory where captioned images will be saved. If omitted, defaults to /tmp/captions/<hash>"`
}

type CaptionItem struct {
	ImgPath string `json:"imgPath" jsonschema:"absolute path to the source image file"`
	Text    string `json:"text" jsonschema:"caption text to display below the image"`
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "compliance-captioner", Version: "v1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "caption",
		Description: "Add caption text below images and save the results",
	}, captionCommand)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func captionCommand(ctx context.Context, req *mcp.CallToolRequest, input CaptionInput) (*mcp.CallToolResult, any, error) {
	totalStart := time.Now()

	if input.OutputPath == "" {
		h := sha256.New()
		for _, c := range input.Captions {
			h.Write([]byte(c.ImgPath))
			h.Write([]byte(c.Text))
		}
		input.OutputPath = filepath.Join("/tmp/captions", hex.EncodeToString(h.Sum(nil))[:12])
	}

	if err := os.MkdirAll(input.OutputPath, 0755); err != nil {
		return nil, nil, fmt.Errorf("create output dir: %w", err)
	}

	fontStart := time.Now()
	fontBytes, err := os.ReadFile("/Library/Fonts/Arial Unicode.ttf")
	if err != nil {
		return nil, nil, fmt.Errorf("read font: %w", err)
	}

	f, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse font: %w", err)
	}
	fontDur := time.Since(fontStart)

	var results []string
	var perImage []string
	for i, item := range input.Captions {
		imgStart := time.Now()
		img, err := constructCaptionedImage(item.ImgPath, item.Text, f)
		if err != nil {
			return nil, nil, fmt.Errorf("caption %d (%s): %w", i+1, item.ImgPath, err)
		}
		constructDur := time.Since(imgStart)

		base := filepath.Base(item.ImgPath)
		outPath := filepath.Join(input.OutputPath, base)

		encodeStart := time.Now()
		out, err := os.Create(outPath)
		if err != nil {
			return nil, nil, fmt.Errorf("create %s: %w", outPath, err)
		}
		if err := png.Encode(out, img); err != nil {
			out.Close()
			return nil, nil, fmt.Errorf("encode %s: %w", outPath, err)
		}
		out.Close()
		encodeDur := time.Since(encodeStart)

		results = append(results, outPath)
		perImage = append(perImage, fmt.Sprintf("  %s: construct=%s encode=%s total=%s", base, constructDur, encodeDur, constructDur+encodeDur))
	}

	totalDur := time.Since(totalStart)
	summary := fmt.Sprintf("Captioned %d images to %s:\n%s\n\nTiming:\n  font load: %s\n%s\n  total: %s",
		len(results), input.OutputPath, strings.Join(results, "\n"),
		fontDur, strings.Join(perImage, "\n"), totalDur)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
		},
	}, nil, nil
}

func constructCaptionedImage(imgPath, text string, f *opentype.Font) (image.Image, error) {
	src, err := os.Open(imgPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", imgPath, err)
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", imgPath, err)
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	ptSize := 16.0

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    ptSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("new face: %w", err)
	}
	defer face.Close()

	lines := wordWrap(text, w-30, face)
	lineH := int(ptSize * 1.5)
	textH := len(lines) * lineH
	padding := 28
	barH := textH + padding*2

	if barH < 100 {
		barH = 100
	}

	newH := h + barH
	dst := image.NewRGBA(image.Rect(0, 0, w, newH))
	draw.Draw(dst, bounds, img, image.Point{}, draw.Src)

	barRect := image.Rect(0, h, w, newH)
	draw.Draw(dst, barRect, &image.Uniform{color.RGBA{0, 0, 0, 235}}, image.Point{}, draw.Src)

	totalTextH := len(lines) * lineH
	textStartY := h + (barH-totalTextH)/2 + lineH - 2

	for li, line := range lines {
		adv := font.MeasureString(face, line)
		textW := adv.Ceil()
		x := (w - textW) / 2
		if x < 10 {
			x = 10
		}
		y := textStartY + li*lineH

		d2 := font.Drawer{
			Dst:  dst,
			Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
			Face: face,
			Dot:  fixed.P(x, y),
		}
		d2.DrawString(line)
	}

	return dst, nil
}

func wordWrap(text string, maxWidth int, face font.Face) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := words[0]

	for _, word := range words[1:] {
		testLine := current + " " + word
		adv := font.MeasureString(face, testLine)
		if adv.Ceil() <= maxWidth {
			current = testLine
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	lines = append(lines, current)
	return lines
}
