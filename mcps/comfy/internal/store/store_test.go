package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const zitFixture = `{
 "1": {"class_type":"CheckpointLoaderSimple","inputs":{"ckpt_name":"{{unet_name}}"}},
 "2": {"class_type":"CLIPTextEncode","inputs":{"text":"{{positive_prompt}}"}},
 "3": {"class_type":"CLIPTextEncode","inputs":{"text":"{{negative_prompt}}"}},
 "4": {"class_type":"EmptyLatentImage","inputs":{"width":"{{width}}","height":"{{height}}"}},
 "5": {"class_type":"EmptyLatentImage","inputs":{"width":"{{width}}","height":"{{height}}"}},
 "6": {"class_type":"LoraLoader","inputs":{"lora_name":"{{lora_name_1}}","strength_model":"{{lora_strength_1}}"}},
 "7": {"class_type":"KSampler","inputs":{"seed":"{{seed}}","steps":"{{steps}}","cfg":"{{cfg}}","denoise":"{{denoise}}"}},
 "8": {"class_type":"SaveImage","inputs":{"filename_prefix":"prefix-{{seed}}","images":["7",0]}}
}`

const inpaintFixture = `{
 "1": {"class_type":"LoadImage","inputs":{"image":"mask.png"}},
 "2": {"class_type":"LoadImage","inputs":{"image":"mask.png"}},
 "3": {"class_type":"LoadImage","inputs":{"image":"ref.png"}},
 "4": {"class_type":"InpaintCropAndStitch","inputs":{"crop_width":"{{crop_width}}","crop_height":"{{crop_height}}","prompt":"{{positive_prompt}}"}},
 "5": {"class_type":"SaveImage","inputs":{"filename_prefix":"out","images":["4",0]}}
}`

func writeFixtures(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func find(t *testing.T, workflows []Workflow, tool string) Workflow {
	t.Helper()
	for _, w := range workflows {
		if w.Tool == tool {
			return w
		}
	}
	t.Fatalf("tool %q not found in %+v", tool, workflows)
	return Workflow{}
}

func spec(w Workflow, name string) (TokenSpec, bool) {
	for _, s := range w.Tokens {
		if s.Name == name {
			return s, true
		}
	}
	return TokenSpec{}, false
}

func TestLoad_ToolsTokensAndImages(t *testing.T) {
	dir := writeFixtures(t, map[string]string{
		"z-image.workflow.json":   zitFixture,
		"an-inpaint.workflow.json": inpaintFixture,
	})

	workflows, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(workflows))
	}

	// files sort alphabetically: an-inpaint first
	if workflows[0].Tool != "an_inpaint" || workflows[1].Tool != "z_image" {
		t.Fatalf("tools = %q, %q", workflows[0].Tool, workflows[1].Tool)
	}

	zit := find(t, workflows, "z_image")
	// duplicate {{width}} in two nodes collapses to one token
	count := 0
	for _, s := range zit.Tokens {
		if s.Name == "width" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("width appeared %d times", count)
	}

	// numeric typing
	for _, name := range []string{"seed", "steps", "cfg", "denoise", "width", "height", "lora_strength_1"} {
		s, ok := spec(zit, name)
		if !ok {
			t.Fatalf("token %s missing", name)
		}
		if s.Type != "number" {
			t.Errorf("token %s: type = %q, want number", name, s.Type)
		}
	}
	for _, name := range []string{"positive_prompt", "unet_name", "lora_name_1", "negative_prompt"} {
		s, ok := spec(zit, name)
		if !ok {
			t.Fatalf("token %s missing", name)
		}
		if s.Type != "string" {
			t.Errorf("token %s: type = %q, want string", name, s.Type)
		}
	}
	// negative_prompt allows empty; positive_prompt must be filled
	if s, _ := spec(zit, "negative_prompt"); !s.AllowEmpty {
		t.Error("negative_prompt should allow empty")
	}
	if s, _ := spec(zit, "positive_prompt"); s.AllowEmpty {
		t.Error("positive_prompt should not allow empty")
	}
	// embedded token in a larger string is still discovered
	if _, ok := spec(zit, "seed"); !ok {
		t.Error("seed token (embedded in filename_prefix) not found")
	}
	if len(zit.Images) != 0 {
		t.Errorf("zit should have no image inputs, got %v", zit.Images)
	}

	// image inputs: mask.png used twice -> one param; ref.png -> another
	inp := find(t, workflows, "an_inpaint")
	if len(inp.Images) != 2 {
		t.Fatalf("expected 2 image inputs, got %+v", inp.Images)
	}
	if inp.Images[0].Param != "mask" || inp.Images[0].Filename != "mask.png" {
		t.Errorf("image[0] = %+v", inp.Images[0])
	}
	if inp.Images[1].Param != "ref" || inp.Images[1].Filename != "ref.png" {
		t.Errorf("image[1] = %+v", inp.Images[1])
	}
	if s, _ := spec(inp, "crop_width"); s.Type != "number" {
		t.Error("crop_width should be numeric")
	}
}

func TestTokenType_NumericPatterns(t *testing.T) {
	cases := map[string]string{
		"lora_strength_1":  "number",
		"lora_strength_12": "number",
		"strength_model_2": "number",
		"strength_clip_3":  "number",
		"lora_name_1":      "string",
		"strength_model":   "string",
		"seed":             "number",
		"fps":              "number",
		"pulid_strength":   "number",
		"positive_prompt":  "string",
	}
	for name, want := range cases {
		if got := tokenType(name); got != want {
			t.Errorf("tokenType(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestToolName(t *testing.T) {
	cases := map[string]string{
		"z-image-turbo-txt2img.workflow.json": "z_image_turbo_txt2img",
		"wan22-5b-ti2v.workflow.json":         "wan22_5b_ti2v",
		"flux2-klein-9b-i2i.workflow.json":    "flux2_klein_9b_i2i",
		"Under_Score.workflow.json":           "Under_Score",
	}
	for in, want := range cases {
		if got := ToolName(in); got != want {
			t.Errorf("ToolName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestImageParamName_DedupAndGuards(t *testing.T) {
	// param collides with a token name -> suffixed
	wf := `{
 "1": {"class_type":"LoadImage","inputs":{"image":"ref.png"}},
 "2": {"class_type":"Foo","inputs":{"x":"{{ref}}"}}
}`
	dir := writeFixtures(t, map[string]string{"x.workflow.json": wf})
	workflows, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if workflows[0].Images[0].Param != "ref_image" {
		t.Errorf("param = %q, want ref_image", workflows[0].Images[0].Param)
	}

	// two filenames with the same stem -> second gets _2
	wf2 := `{
 "1": {"class_type":"LoadImage","inputs":{"image":"ref.png"}},
 "2": {"class_type":"LoadImage","inputs":{"image":"ref.jpg"}},
 "3": {"class_type":"Foo","inputs":{"x":"{{a}}"}}
}`
	dir2 := writeFixtures(t, map[string]string{"y.workflow.json": wf2})
	ws2, err := Load(dir2)
	if err != nil {
		t.Fatal(err)
	}
	params := []string{ws2[0].Images[0].Param, ws2[0].Images[1].Param}
	joined := strings.Join(params, ",")
	if joined != "ref,ref_2" {
		t.Errorf("params = %s, want ref,ref_2", joined)
	}
}

func TestLoad_Errors(t *testing.T) {
	t.Run("missing dir", func(t *testing.T) {
		if _, err := Load(filepath.Join(t.TempDir(), "nope")); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("empty dir", func(t *testing.T) {
		if _, err := Load(t.TempDir()); err == nil || !strings.Contains(err.Error(), "no *.workflow.json") {
			t.Fatalf("expected empty-dir error, got %v", err)
		}
	})
	t.Run("bad json", func(t *testing.T) {
		dir := writeFixtures(t, map[string]string{"x.workflow.json": "{not json"})
		if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "parse") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
	t.Run("no tokens", func(t *testing.T) {
		dir := writeFixtures(t, map[string]string{
			"x.workflow.json": `{"1":{"class_type":"SaveImage","inputs":{"filename_prefix":"a"}}}`,
		})
		if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "no {{token}}") {
			t.Fatalf("expected tokenless error, got %v", err)
		}
	})
	t.Run("tool name collision", func(t *testing.T) {
		dir := writeFixtures(t, map[string]string{
			"a-b.workflow.json": strings.ReplaceAll(zitFixture, "{{", "{{"),
			"a_b.workflow.json": zitFixture,
		})
		if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "collision") {
			t.Fatalf("expected collision error, got %v", err)
		}
	})
}
