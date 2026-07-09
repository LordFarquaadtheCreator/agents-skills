package generate

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fogleman/gg"
	"github.com/go-pdf/fpdf"
	"github.com/golang/freetype/truetype"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
)

// Input is the agent-facing generate_story_pdf tool input.
type Input struct {
	Title            string  `json:"title" jsonschema:"required,Story title shown in the page footer and used as the output directory name"`
	Pages            []Page  `json:"pages" jsonschema:"required,Array of pages — each has an image file path and text"`
	OutputDir        string  `json:"outputDir,omitempty" jsonschema:"Base directory for output. A subdirectory named after the title is created inside. Defaults to ~/Desktop."`
	FontSize         float64 `json:"fontSize,omitempty" jsonschema:"Max body font size in points. Binary-searched down to fit text on page. Defaults to 30."`
	LightenFactor    float64 `json:"lightenFactor,omitempty" jsonschema:"How muted the background is (0.0=original, 1.0=white). Defaults to 0.8."`
	PreviewAfterPage int     `json:"previewAfterPage,omitempty" jsonschema:"Pages after this number are blurred in the preview PDF. E.g. 3 means pages 1-3 are clear, 4+ are blurred. 0 or omitted means no preview generated."`
}

// Page is one page of the story.
type Page struct {
	Image string `json:"image" jsonschema:"required,Absolute file path to a PNG or JPEG image"`
	Text  string `json:"text" jsonschema:"required,Story text for this page. Markdown supported: **bold**, *italic*, \n for line breaks, \n\n for paragraph breaks."`
}

// Output is returned to the agent.
type Output struct {
	OutputDir       string   `json:"outputDir"`
	PDFPath         string   `json:"pdfPath"`
	PNGPaths        []string `json:"pngPaths"`
	PageCount       int      `json:"pageCount"`
	PreviewPDFPath  string   `json:"previewPdfPath,omitempty"`
	PreviewPNGPaths []string `json:"previewPngPaths,omitempty"`
}

const (
	scale                = 2.0
	pageWPt              = 1200.0
	pageHPt              = 600.0
	pageWPx              = pageWPt * scale
	pageHPx              = pageHPt * scale
	targetImgHeight      = pageHPx
	textPad              = 20.0 * scale
	footerHeight         = 24.0 * scale
	defaultFontSize      = 30.0 * scale
	defaultLightenFactor = 0.80
	minFontSize          = 6.0 * scale
	previewBlurSigma     = 25.0

	fontRegular = "/System/Library/Fonts/Supplemental/AppleMyungjo.ttf"
	fontBold    = "/System/Library/Fonts/Supplemental/AppleMyungjo.ttf"
	fontItalic  = "/System/Library/Fonts/Supplemental/AppleMyungjo.ttf"
)

type fontFamily struct {
	regular *truetype.Font
	bold    *truetype.Font
	italic  *truetype.Font
}

func loadFontFamily() (*fontFamily, error) {
	rBytes, err := os.ReadFile(fontRegular)
	if err != nil {
		return nil, fmt.Errorf("load regular font: %w", err)
	}
	bBytes, err := os.ReadFile(fontBold)
	if err != nil {
		return nil, fmt.Errorf("load bold font: %w", err)
	}
	iBytes, err := os.ReadFile(fontItalic)
	if err != nil {
		return nil, fmt.Errorf("load italic font: %w", err)
	}
	rFont, err := truetype.Parse(rBytes)
	if err != nil {
		return nil, fmt.Errorf("parse regular font: %w", err)
	}
	bFont, err := truetype.Parse(bBytes)
	if err != nil {
		return nil, fmt.Errorf("parse bold font: %w", err)
	}
	iFont, err := truetype.Parse(iBytes)
	if err != nil {
		return nil, fmt.Errorf("parse italic font: %w", err)
	}
	return &fontFamily{regular: rFont, bold: bFont, italic: iFont}, nil
}

func (ff *fontFamily) face(forFace string, size float64) font.Face {
	opt := &truetype.Options{Size: size, DPI: 72, Hinting: font.HintingFull}
	switch forFace {
	case "B":
		return truetype.NewFace(ff.bold, opt)
	case "I":
		return truetype.NewFace(ff.italic, opt)
	default:
		return truetype.NewFace(ff.regular, opt)
	}
}

func measureWith(f font.Face, s string) float64 {
	d := &font.Drawer{Face: f}
	return float64(d.MeasureString(s)) / 64
}

// Run builds the PDF and PNGs from the given input and writes them to disk.
func Run(args Input) (Output, error) {
	if len(args.Pages) == 0 {
		return Output{}, fmt.Errorf("at least one page is required")
	}

	fontSize := args.FontSize * scale
	if fontSize <= 0 {
		fontSize = defaultFontSize
	}
	lighten := args.LightenFactor
	if lighten <= 0 {
		lighten = defaultLightenFactor
	}

	ff, err := loadFontFamily()
	if err != nil {
		return Output{}, fmt.Errorf("load fonts: %w", err)
	}

	baseDir := args.OutputDir
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Output{}, fmt.Errorf("resolve home: %w", err)
		}
		baseDir = filepath.Join(home, "Desktop")
	}

	name := sanitizeFilename(strings.TrimSpace(args.Title))
	if name == "" {
		name = "Story"
	}
	outDir, err := resolveOutputDir(baseDir, name)
	if err != nil {
		return Output{}, err
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return Output{}, fmt.Errorf("create output dir: %w", err)
	}

	pdfPath := filepath.Join(outDir, name+".pdf")
	pngPaths := make([]string, len(args.Pages))
	pageImgs := make([]image.Image, len(args.Pages))

	pdf := fpdf.New("", "pt", "", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(0, 0, 0)
	pdf.SetTitle(args.Title, false)
	pdf.SetAuthor("senor", false)
	pdf.SetCreator("create-story MCP", false)

	log.Printf("create-story: generating %d-page PDF, title=%q, fontSize=%.1f, lighten=%.2f, outDir=%s",
		len(args.Pages), args.Title, fontSize, lighten, outDir)

	for i, page := range args.Pages {
		log.Printf("create-story: page %d/%d — loading image %s", i+1, len(args.Pages), page.Image)
		img, err := loadImage(page.Image)
		if err != nil {
			return Output{}, fmt.Errorf("page %d: %w", i+1, err)
		}
		log.Printf("create-story: page %d — rendering", i+1)
		pageImg := renderPageImage(ff, img,
			sanitizeASCII(page.Text),
			sanitizeASCII(args.Title),
			i+1, fontSize, lighten)
		pageImgs[i] = pageImg

		pngPath := filepath.Join(outDir, fmt.Sprintf("%s.%d.png", name, i+1))
		if err := savePNG(pageImg, pngPath); err != nil {
			return Output{}, fmt.Errorf("page %d: save png: %w", i+1, err)
		}
		pngPaths[i] = pngPath
		log.Printf("create-story: page %d — saved %s", i+1, pngPath)

		pdf.AddPageFormat("", fpdf.SizeType{Wd: pageWPt, Ht: pageHPt})
		pdf.ImageOptions(pngPath, 0, 0, pageWPt, pageHPt, false,
			fpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	}

	log.Printf("create-story: writing PDF to %s", pdfPath)
	if err := pdf.OutputFileAndClose(pdfPath); err != nil {
		return Output{}, fmt.Errorf("write pdf: %w", err)
	}
	log.Printf("create-story: done — %d pages, %s", len(args.Pages), pdfPath)

	previewPDFPath := ""
	var previewPNGPaths []string

	if args.PreviewAfterPage > 0 && args.PreviewAfterPage < len(args.Pages) {
		previewDir := filepath.Join(outDir, "preview")
		if err := os.MkdirAll(previewDir, 0755); err != nil {
			return Output{}, fmt.Errorf("create preview dir: %w", err)
		}

		previewPDFPath = filepath.Join(previewDir, name+"_preview.pdf")
		previewPNGPaths = make([]string, len(args.Pages))

		previewPDF := fpdf.New("", "pt", "", "")
		previewPDF.SetAutoPageBreak(false, 0)
		previewPDF.SetMargins(0, 0, 0)
		previewPDF.SetTitle(args.Title+" (Preview)", false)
		previewPDF.SetAuthor("senor", false)
		previewPDF.SetCreator("create-story MCP", false)

		log.Printf("create-story: generating preview PDF, blur after page %d", args.PreviewAfterPage)

		for i, pageImg := range pageImgs {
			var img image.Image
			if i+1 > args.PreviewAfterPage {
				img = gaussianBlur(pageImg, previewBlurSigma)
			} else {
				img = pageImg
			}

			pngPath := filepath.Join(previewDir, fmt.Sprintf("%s_preview_%d.png", name, i+1))
			if err := savePNG(img, pngPath); err != nil {
				return Output{}, fmt.Errorf("preview page %d: save png: %w", i+1, err)
			}
			previewPNGPaths[i] = pngPath

			previewPDF.AddPageFormat("", fpdf.SizeType{Wd: pageWPt, Ht: pageHPt})
			previewPDF.ImageOptions(pngPath, 0, 0, pageWPt, pageHPt, false,
				fpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
		}

		log.Printf("create-story: writing preview PDF to %s", previewPDFPath)
		if err := previewPDF.OutputFileAndClose(previewPDFPath); err != nil {
			return Output{}, fmt.Errorf("write preview pdf: %w", err)
		}
		log.Printf("create-story: preview done — %s", previewPDFPath)
	}

	return Output{
		OutputDir:       outDir,
		PDFPath:         pdfPath,
		PNGPaths:        pngPaths,
		PageCount:       len(args.Pages),
		PreviewPDFPath:  previewPDFPath,
		PreviewPNGPaths: previewPNGPaths,
	}, nil
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return img, nil
}

func savePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

type pageItem struct {
	word, face    string
	isParaBreak   bool
	isLineBreak   bool
	noSpaceBefore bool
}

func flattenWords(story string) []pageItem {
	var all []pageItem
	for _, block := range strings.Split(story, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		for _, ln := range strings.Split(block, "\n") {
			if ln == "" {
				all = append(all, pageItem{"", "", false, true, false})
				continue
			}
			segs := parseMarkdown(ln)
			for segIdx, s := range segs {
				words := strings.Fields(s.text)
				if len(words) == 0 {
					continue
				}
				noSpace := segIdx > 0 &&
					!strings.HasSuffix(segs[segIdx-1].text, " ") &&
					!strings.HasPrefix(s.text, " ")
				for wIdx, w := range words {
					pi := pageItem{w, s.face, false, false, false}
					if noSpace && wIdx == 0 {
						pi.noSpaceBefore = true
					}
					all = append(all, pi)
				}
			}
			all = append(all, pageItem{"", "", false, true, false})
		}
		if len(all) > 0 && all[len(all)-1].isLineBreak {
			all = all[:len(all)-1]
		}
		all = append(all, pageItem{"", "", true, false, false})
	}
	return all
}

func renderPageImage(ff *fontFamily, srcImg image.Image, story, title string, pageNum int, maxFontSize, lighten float64) image.Image {
	srcBounds := srcImg.Bounds()
	srcW, srcH := float64(srcBounds.Dx()), float64(srcBounds.Dy())
	imgScale := targetImgHeight / srcH
	renderW := srcW * imgScale
	renderH := targetImgHeight
	if renderW > pageWPx/2 {
		imgScale = (pageWPx / 2) / srcW
		renderW = pageWPx / 2
		renderH = srcH * imgScale
	}

	r, g, b := getDominantColorFromImage(srcImg, lighten)

	dc := gg.NewContext(int(pageWPx), int(pageHPx))
	dc.SetRGB(float64(r)/255, float64(g)/255, float64(b)/255)
	dc.Clear()

	scaledW, scaledH := int(renderW), int(renderH)
	scaled := image.NewRGBA(image.Rect(0, 0, scaledW, scaledH))
	xdraw.BiLinear.Scale(scaled, scaled.Bounds(), srcImg, srcBounds, xdraw.Over, nil)
	dc.DrawImage(scaled, 0, 0)

	textX := renderW + textPad
	maxW := pageWPx - textX - textPad
	availH := pageHPx - textPad*2 - footerHeight

	all := flattenWords(story)

	// binary search font size
	minSize := minFontSize
	maxSize := maxFontSize
	bestSize := minSize

	for minSize <= maxSize {
		try := (minSize + maxSize) / 2
		lh := try * 1.28
		paraPad := lh * 0.4
		need := 0.0
		lineW := 0.0
		prevWord := false

		for _, it := range all {
			if it.isParaBreak {
				if prevWord {
					need += lh
					lineW = 0
					prevWord = false
				}
				need += paraPad
				continue
			}
			if it.isLineBreak {
				if prevWord {
					need += lh
					lineW = 0
					prevWord = false
				}
				continue
			}
			f := ff.face(it.face, try)
			wW := measureWith(f, it.word)
			testW := lineW
			if prevWord && !it.noSpaceBefore {
				testW += measureWith(f, " ")
			}
			testW += wW
			if testW > maxW && prevWord {
				need += lh
				if need > availH {
					break
				}
				lineW = wW
				prevWord = true
			} else {
				lineW = testW
				prevWord = true
			}
		}
		if prevWord {
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

	faceReg := ff.face("", fontSize)
	faceBld := ff.face("B", fontSize)
	faceItl := ff.face("I", fontSize)
	faceFor := func(s string) font.Face {
		switch s {
		case "B":
			return faceBld
		case "I":
			return faceItl
		default:
			return faceReg
		}
	}

	// render text
	dc.SetRGB(26.0/255, 26.0/255, 26.0/255)
	ty := textPad + fontSize
	skipPad := true
	var lineWords []pageItem
	lineW := 0.0

	flush := func() {
		if len(lineWords) == 0 {
			return
		}
		x := textX
		for j, w := range lineWords {
			f := faceFor(w.face)
			dc.SetFontFace(f)
			dc.DrawString(w.word, x, ty)
			x += measureWith(f, w.word)
			if j < len(lineWords)-1 && !lineWords[j+1].noSpaceBefore {
				x += measureWith(f, " ")
			}
		}
		ty += lineH
		lineWords = nil
		lineW = 0
	}

	for _, it := range all {
		if it.isParaBreak {
			flush()
			if !skipPad {
				ty += paraPad
			}
			skipPad = false
			continue
		}
		if it.isLineBreak {
			flush()
			skipPad = false
			continue
		}
		skipPad = false
		f := faceFor(it.face)
		wW := measureWith(f, it.word)
		testW := lineW
		if len(lineWords) > 0 && !it.noSpaceBefore {
			testW += measureWith(f, " ")
		}
		testW += wW
		if testW > maxW && len(lineWords) > 0 {
			flush()
			lineWords = append(lineWords, it)
			lineW = wW
		} else {
			lineWords = append(lineWords, it)
			lineW = testW
		}
	}
	flush()

	// footer — fixed size, not scaled with body text
	footerSize := 14.0 * scale
	footerFace := ff.face("", footerSize)
	label := fmt.Sprintf("%s #%d", title, pageNum)
	dc.SetFontFace(footerFace)
	dc.SetRGB(100.0/255, 100.0/255, 100.0/255)
	lw := measureWith(footerFace, label)
	dc.DrawString(label, textX+maxW-lw, pageHPx-textPad)

	return dc.Image()
}

// getDominantColorFromImage extracts a muted background color from the image.
func getDominantColorFromImage(img image.Image, lighten float64) (int, int, int) {
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

type segment struct{ text, face string }

// parseMarkdown splits text into segments with face info.
// Supports **bold** and *italic* inline markup.
func parseMarkdown(s string) []segment {
	var segs []segment
	for s != "" {
		if strings.HasPrefix(s, "**") {
			end := strings.Index(s[2:], "**")
			if end < 0 {
				break
			}
			text := s[2 : 2+end]
			if text != "" {
				segs = append(segs, segment{text, "B"})
			}
			s = s[2+end+2:]
			continue
		}
		if strings.HasPrefix(s, "*") {
			end := strings.Index(s[1:], "*")
			if end < 0 {
				break
			}
			text := s[1 : 1+end]
			if text != "" {
				segs = append(segs, segment{text, "I"})
			}
			s = s[1+end+1:]
			continue
		}
		bi := strings.Index(s, "**")
		ii := strings.Index(s, "*")
		var next int
		if bi < 0 && ii < 0 {
			next = -1
		} else if bi < 0 {
			next = ii
		} else if ii < 0 {
			next = bi
		} else {
			next = min(bi, ii)
		}
		if next < 0 {
			if s != "" {
				segs = append(segs, segment{s, ""})
			}
			break
		}
		if next > 0 {
			segs = append(segs, segment{s[:next], ""})
		}
		s = s[next:]
	}
	return segs
}

// sanitizeASCII replaces common Unicode punctuation with ASCII equivalents
// and strips any remaining non-ASCII characters.
func sanitizeASCII(s string) string {
	s = strings.ReplaceAll(s, "\u2014", "-")
	s = strings.ReplaceAll(s, "\u2013", "-")
	s = strings.ReplaceAll(s, "\u201C", `"`)
	s = strings.ReplaceAll(s, "\u201D", `"`)
	s = strings.ReplaceAll(s, "\u2018", "'")
	s = strings.ReplaceAll(s, "\u2019", "'")
	s = strings.ReplaceAll(s, "\u2026", "...")
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < 128 {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func sanitizeFilename(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '/' || r == ':' || r == '\\' || r == '\x00' {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func resolveOutputDir(base, name string) (string, error) {
	dir := filepath.Join(base, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return dir, nil
	}
	for i := 2; i < 100; i++ {
		candidate := filepath.Join(base, fmt.Sprintf("%s %d", name, i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not find available output directory after 100 attempts")
}

// gaussianBlur applies a separable Gaussian blur to the source image.
func gaussianBlur(src image.Image, sigma float64) image.Image {
	if sigma <= 0 {
		return src
	}
	radius := int(math.Ceil(sigma * 3))
	if radius < 1 {
		radius = 1
	}
	size := radius*2 + 1
	kernel := make([]float64, size)
	var sum float64
	for i := 0; i < size; i++ {
		x := float64(i - radius)
		v := math.Exp(-(x * x) / (2 * sigma * sigma))
		kernel[i] = v
		sum += v
	}
	for i := range kernel {
		kernel[i] /= sum
	}

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			rgba.SetRGBA(x, y, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)})
		}
	}

	// horizontal pass
	tmp := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, b, a float64
			for k := -radius; k <= radius; k++ {
				sx := x + k
				if sx < 0 {
					sx = 0
				}
				if sx >= w {
					sx = w - 1
				}
				weight := kernel[k+radius]
				px := rgba.RGBAAt(sx, y)
				r += float64(px.R) * weight
				g += float64(px.G) * weight
				b += float64(px.B) * weight
				a += float64(px.A) * weight
			}
			tmp.SetRGBA(x, y, color.RGBA{
				R: uint8(math.Round(r)),
				G: uint8(math.Round(g)),
				B: uint8(math.Round(b)),
				A: uint8(math.Round(a)),
			})
		}
	}

	// vertical pass
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, b, a float64
			for k := -radius; k <= radius; k++ {
				sy := y + k
				if sy < 0 {
					sy = 0
				}
				if sy >= h {
					sy = h - 1
				}
				weight := kernel[k+radius]
				px := tmp.RGBAAt(x, sy)
				r += float64(px.R) * weight
				g += float64(px.G) * weight
				b += float64(px.B) * weight
				a += float64(px.A) * weight
			}
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(math.Round(r)),
				G: uint8(math.Round(g)),
				B: uint8(math.Round(b)),
				A: uint8(math.Round(a)),
			})
		}
	}

	return dst
}
