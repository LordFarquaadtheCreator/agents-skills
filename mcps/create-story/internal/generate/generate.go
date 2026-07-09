package generate

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-pdf/fpdf"
)

// Input is the agent-facing generate_story_pdf tool input.
type Input struct {
	Title         string  `json:"title" jsonschema:"required,Story title shown in the page footer"`
	Pages         []Page  `json:"pages" jsonschema:"required,Array of pages — each has a base64 image and text"`
	OutputDir     string  `json:"outputDir,omitempty" jsonschema:"Directory to save the PDF. Defaults to ~/Downloads."`
	Filename      string  `json:"filename,omitempty" jsonschema:"PDF filename. Defaults to <title>.pdf."`
	FontSize      float64 `json:"fontSize,omitempty" jsonschema:"Max body font size in points. Binary-searched down to fit text on page. Defaults to 30."`
	LightenFactor float64 `json:"lightenFactor,omitempty" jsonschema:"How muted the background is (0.0=original, 1.0=white). Defaults to 0.8."`
}

// Page is one page of the story.
type Page struct {
	Image string `json:"image" jsonschema:"required,Base64-encoded PNG or JPEG image data (without data URI prefix)"`
	Text  string `json:"text" jsonschema:"required,Story text for this page. Supports <b>bold</b>, <i>italic</i>, and <br> for line breaks."`
}

// Output is returned to the agent.
type Output struct {
	OutputPath string `json:"outputPath"`
	PageCount  int    `json:"pageCount"`
	Filename   string `json:"filename"`
}

const (
	defaultFontSize      = 30.0
	defaultLightenFactor = 0.80
	targetImgHeight      = 600.0
	textPad              = 20.0
)

// Run builds the PDF from the given input and writes it to disk.
func Run(args Input) (Output, error) {
	if len(args.Pages) == 0 {
		return Output{}, fmt.Errorf("at least one page is required")
	}

	fontSize := args.FontSize
	if fontSize <= 0 {
		fontSize = defaultFontSize
	}
	lighten := args.LightenFactor
	if lighten <= 0 {
		lighten = defaultLightenFactor
	}

	dir := args.OutputDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Output{}, fmt.Errorf("resolve output dir: %w", err)
		}
		dir = filepath.Join(home, "Downloads")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Output{}, fmt.Errorf("create output dir: %w", err)
	}

	fname := args.Filename
	if fname == "" {
		fname = strings.TrimSpace(args.Title)
		if fname == "" {
			fname = "Story"
		}
		fname += ".pdf"
	}

	pdf := fpdf.New("", "pt", "", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(0, 0, 0)

	for i, page := range args.Pages {
		imgPath, imgType, err := decodeBase64Image(page.Image, i)
		if err != nil {
			return Output{}, fmt.Errorf("page %d: %w", i+1, err)
		}
		renderPage(pdf, imgPath, imgType, sanitizeASCII(page.Text), sanitizeASCII(args.Title), i+1, fontSize, lighten)
		os.Remove(imgPath)
	}

	outPath := filepath.Join(dir, fname)
	if err := pdf.OutputFileAndClose(outPath); err != nil {
		return Output{}, fmt.Errorf("write pdf: %w", err)
	}

	return Output{
		OutputPath: outPath,
		PageCount:  len(args.Pages),
		Filename:   fname,
	}, nil
}

// decodeBase64Image decodes a base64 string to a temp file and returns its path and image type.
func decodeBase64Image(b64 string, idx int) (string, string, error) {
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return "", "", fmt.Errorf("decode base64: %w", err)
	}

	imgType := "PNG"
	if isJPEG(data) {
		imgType = "JPG"
	}

	ext := ".png"
	if imgType == "JPG" {
		ext = ".jpg"
	}

	tmp, err := os.CreateTemp("", fmt.Sprintf("story-page-%d-*%s", idx+1, ext))
	if err != nil {
		return "", "", fmt.Errorf("create temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", "", fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()
	return tmp.Name(), imgType, nil
}

func isJPEG(data []byte) bool {
	return len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF
}

// getDominantColor extracts a muted background color from the image.
func getDominantColor(path string, lighten float64) (int, int, int) {
	f, err := os.Open(path)
	if err != nil {
		return 240, 235, 225
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return 240, 235, 225
	}

	b := img.Bounds()
	type bucket struct{ r, g, bl, count int }
	pal := make(map[int]*bucket)
	const step = 16
	const res = 48

	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			pr, pg, pb, _ := img.At(x, y).RGBA()
			r := int(pr>>8) / res * res
			g := int(pg>>8) / res * res
			bl := int(pb>>8) / res * res
			key := r<<16 | g<<8 | bl
			if p, ok := pal[key]; ok {
				p.count++
			} else {
				pal[key] = &bucket{r, g, bl, 1}
			}
		}
	}

	if len(pal) == 0 {
		return 240, 235, 225
	}

	var list []*bucket
	for _, v := range pal {
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].count > list[j].count })

	for _, p := range list {
		if p.count < 5 {
			continue
		}
		lum := (p.r + p.g + p.bl) / 3
		if lum < 50 || lum > 220 {
			continue
		}
		mx := max(p.r, max(p.g, p.bl))
		mn := min(p.r, min(p.g, p.bl))
		sat := 0
		if mx > 0 {
			sat = (mx - mn) * 255 / mx
		}
		if sat > 30 {
			return lightenColor(p.r, p.g, p.bl, lighten)
		}
	}

	for _, p := range list {
		lum := (p.r + p.g + p.bl) / 3
		if lum > 30 && lum < 230 {
			return lightenColor(p.r, p.g, p.bl, lighten)
		}
	}
	return 240, 235, 225
}

func lightenColor(r, g, bl int, lighten float64) (int, int, int) {
	r = int(math.Min(255, float64(r)+(255-float64(r))*lighten))
	g = int(math.Min(255, float64(g)+(255-float64(g))*lighten))
	bl = int(math.Min(255, float64(bl)+(255-float64(bl))*lighten))
	return r, g, bl
}

func imgSize(path string) (int, int) {
	f, err := os.Open(path)
	if err != nil {
		return 720, 1024
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return 720, 1024
	}
	b := img.Bounds()
	return b.Dx(), b.Dy()
}

type segment struct{ text, face string }

func parseHTML(s string) []segment {
	var segs []segment
	cur := s
	for cur != "" {
		ei := strings.Index(cur, "<")
		if ei < 0 {
			if t := strings.TrimSpace(stripTags(cur)); t != "" {
				segs = append(segs, segment{t, ""})
			}
			break
		}
		if ei > 0 {
			if t := strings.TrimSpace(stripTags(cur[:ei])); t != "" {
				segs = append(segs, segment{t, ""})
			}
			cur = cur[ei:]
			continue
		}
		if strings.HasPrefix(cur, "<br>") || strings.HasPrefix(cur, "<br/>") {
			cur = cur[4:]
			if strings.HasPrefix(cur, "/>") {
				cur = cur[1:]
			}
			continue
		}
		if strings.HasPrefix(cur, "</") {
			cur = cur[4+len(string(cur[2])):]
			continue
		}
		if strings.HasPrefix(cur, "<b>") {
			end := strings.Index(cur, "</b>")
			if end < 0 {
				continue
			}
			if t := strings.TrimSpace(stripTags(cur[3:end])); t != "" {
				segs = append(segs, segment{t, "B"})
			}
			cur = cur[end+4:]
			continue
		}
		if strings.HasPrefix(cur, "<i>") {
			end := strings.Index(cur, "</i>")
			if end < 0 {
				continue
			}
			if t := strings.TrimSpace(stripTags(cur[3:end])); t != "" {
				segs = append(segs, segment{t, "I"})
			}
			cur = cur[end+4:]
			continue
		}
		if n := strings.Index(cur, ">"); n >= 0 {
			cur = cur[n+1:]
		} else {
			break
		}
	}
	return segs
}

// sanitizeASCII replaces common Unicode punctuation with ASCII equivalents
// and strips any remaining non-ASCII characters so fpdf core fonts can render them.
func sanitizeASCII(s string) string {
	s = strings.ReplaceAll(s, "\u2014", "-")   // em dash
	s = strings.ReplaceAll(s, "\u2013", "-")   // en dash
	s = strings.ReplaceAll(s, "\u201C", `"`)   // left double quote
	s = strings.ReplaceAll(s, "\u201D", `"`)   // right double quote
	s = strings.ReplaceAll(s, "\u2018", "'")   // left single quote
	s = strings.ReplaceAll(s, "\u2019", "'")   // right single quote
	s = strings.ReplaceAll(s, "\u2026", "...") // ellipsis
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < 128 {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func stripTags(s string) string {
	s = strings.ReplaceAll(s, "<b>", "")
	s = strings.ReplaceAll(s, "</b>", "")
	s = strings.ReplaceAll(s, "<i>", "")
	s = strings.ReplaceAll(s, "</i>", "")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	return s
}

func renderPage(pdf *fpdf.Fpdf, imgPath, imgType, story, title string, pageNum int, maxFontSize, lighten float64) {
	iw, ih := imgSize(imgPath)
	scale := targetImgHeight / float64(ih)
	renderW := float64(iw) * scale
	renderH := targetImgHeight

	pH := renderH
	pW := pH * 2

	r, g, bl := getDominantColor(imgPath, lighten)

	size := fpdf.SizeType{Wd: pW, Ht: pH}
	pdf.AddPageFormat("", size)
	pdf.SetFillColor(r, g, bl)
	pdf.Rect(0, 0, pW, pH, "F")

	pdf.ImageOptions(imgPath, 0, 0, renderW, renderH, false,
		fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, 0, "")

	tx := renderW + textPad
	maxW := pW - tx - textPad
	availH := pH - textPad*2 - 24 // 24 = footer height

	// Flatten all blocks into word list with paragraph breaks
	type item struct {
		word, face  string
		isParaBreak bool
	}
	var all []item
	for _, block := range strings.Split(story, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		segs := parseHTML(block)
		if len(segs) == 0 {
			continue
		}
		for _, s := range segs {
			for _, w := range strings.Fields(s.text) {
				all = append(all, item{w, s.face, false})
			}
		}
		all = append(all, item{"", "", true})
	}

	// Binary search for best font size that fits
	measure := fpdf.New("", "pt", "", "")
	measure.SetAutoPageBreak(false, 0)
	measure.AddPage()
	minSize := 6.0
	maxSize := maxFontSize
	bestSize := minSize

	for minSize <= maxSize {
		try := (minSize + maxSize) / 2
		lh := try * 1.28
		paraPad := lh * 0.4
		need := 0.0
		line := ""
		for _, it := range all {
			if it.isParaBreak {
				if line != "" {
					need += lh
					line = ""
				}
				need += paraPad
				continue
			}
			measure.SetFont("Helvetica", it.face, try)
			test := line
			if test != "" {
				test += " "
			}
			test += it.word
			if measure.GetStringWidth(test) > maxW && line != "" {
				need += lh
				if need > availH {
					break
				}
				line = it.word
			} else {
				line = test
			}
		}
		if line != "" {
			need += lh
		}
		if need <= availH {
			bestSize = try
			minSize = try + 0.5
		} else {
			maxSize = try - 0.5
		}
	}

	fontSize := math.Min(bestSize, maxFontSize)
	lineH := fontSize * 1.28
	paraPad := lineH * 0.4

	// Render text
	ty := textPad
	pdf.SetTextColor(26, 26, 26)
	line := ""
	lface := ""
	skipPad := true

	for _, it := range all {
		if it.isParaBreak {
			if line != "" {
				pdf.SetFont("Helvetica", lface, fontSize)
				pdf.SetXY(tx, ty)
				pdf.CellFormat(maxW, lineH, line, "", 0, "L", false, 0, "")
				ty += lineH
				line = ""
				lface = ""
			}
			if !skipPad {
				ty += paraPad
			}
			skipPad = false
			continue
		}
		skipPad = false
		test := line
		if test != "" {
			test += " "
		}
		test += it.word
		pdf.SetFont("Helvetica", it.face, fontSize)
		if pdf.GetStringWidth(test) > maxW && line != "" {
			pdf.SetFont("Helvetica", lface, fontSize)
			pdf.SetXY(tx, ty)
			pdf.CellFormat(maxW, lineH, line, "", 0, "L", false, 0, "")
			ty += lineH
			line = it.word
			lface = it.face
		} else {
			line = test
			lface = it.face
		}
	}
	if line != "" {
		pdf.SetFont("Helvetica", lface, fontSize)
		pdf.SetXY(tx, ty)
		pdf.CellFormat(maxW, lineH, line, "", 0, "L", false, 0, "")
	}

	// Footer
	label := fmt.Sprintf("%s - %d", title, pageNum)
	pdf.SetFont("Helvetica", "", fontSize)
	pdf.SetTextColor(100, 100, 100)
	fw := pdf.GetStringWidth(label)
	pdf.SetXY(tx+maxW-fw, pH-textPad-fontSize)
	pdf.CellFormat(fw, fontSize, label, "", 0, "R", false, 0, "")
}
