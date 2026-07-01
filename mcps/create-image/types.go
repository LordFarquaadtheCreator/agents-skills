package main

// Request is the payload sent to the Modal ComfyUI API.
type Request struct {
	PositivePrompt string `json:"positive_prompt"`
	NegativePrompt string `json:"negative_prompt"`

	Seed        *int64   `json:"seed,omitempty"`
	Steps       *int     `json:"steps,omitempty"`
	Width       *int     `json:"width,omitempty"`
	Height      *int     `json:"height,omitempty"`
	Cfg         *float64 `json:"cfg,omitempty"`
	SamplerName *string  `json:"sampler_name,omitempty"`
	Scheduler   *string  `json:"scheduler,omitempty"`
	Denoise     *float64 `json:"denoise,omitempty"`

	UnetName string `json:"unet_name,omitempty"`

	LoraName1     string  `json:"lora_name_1"`
	LoraStrength1 float64 `json:"lora_strength_1"`
	LoraName2     string  `json:"lora_name_2"`
	LoraStrength2 float64 `json:"lora_strength_2"`
	LoraName3     string  `json:"lora_name_3"`
	LoraStrength3 float64 `json:"lora_strength_3"`
}

// ModelCard is the top-level structure of model_card.yaml.
type ModelCard struct {
	Defaults   Defaults    `yaml:"defaults"`
	BaseModels []BaseModel `yaml:"base_models"`
	Loras      []LoraEntry `yaml:"loras"`
}

type Defaults struct {
	Cfg       float64 `yaml:"cfg"`
	Steps     int     `yaml:"steps"`
	Width     int     `yaml:"width"`
	Height    int     `yaml:"height"`
	Sampler   string  `yaml:"sampler"`
	Scheduler string  `yaml:"scheduler"`
}

type BaseModel struct {
	Filename     string   `yaml:"filename"`
	ModelName    string   `yaml:"model_name"`
	Type         string   `yaml:"type"`
	Architecture string   `yaml:"architecture"`
	Link         string   `yaml:"link"`
	PromptStyle  string   `yaml:"prompt_style"`
	Sampler      string   `yaml:"sampler"`
	Cfg          float64  `yaml:"cfg"`
	Steps        int      `yaml:"steps"`
	Width        int      `yaml:"width"`
	Height       int      `yaml:"height"`
	Keywords     []string `yaml:"keywords"`
	Notes        string   `yaml:"notes"`
}

type LoraEntry struct {
	Filename            string   `yaml:"filename"`
	Name                string   `yaml:"name"`
	Type                string   `yaml:"type"`
	PromptStyle         string   `yaml:"prompt_style"`
	Sampler             string   `yaml:"sampler"`
	Link                string   `yaml:"link"`
	RecommendedStrength float64  `yaml:"reccomended_strength"`
	Keywords            []string `yaml:"keywords"`
	Notes               string   `yaml:"notes"`
}

// GenerateImageInput is the typed input for the generate_image MCP tool.
type GenerateImageInput struct {
	PositivePrompt string  `json:"positive_prompt" jsonschema:"required,description=The positive prompt describing what to generate"`
	NegativePrompt string  `json:"negative_prompt" jsonschema:"description=The negative prompt (default: empty)"`
	Seed           *int64  `json:"seed" jsonschema:"description=Random seed for reproducibility (optional, API default if omitted)"`
	Steps          *int    `json:"steps" jsonschema:"description=Number of sampling steps (default: 16)"`
	Width          *int    `json:"width" jsonschema:"description=Image width in pixels (default: 720)"`
	Height         *int    `json:"height" jsonschema:"description=Image height in pixels (default: 1024)"`
	Repeat         int     `json:"repeat" jsonschema:"description=Number of times to generate with incrementing seed (default: 1)"`
	LoraFilename1  string  `json:"lora_filename_1" jsonschema:"required,description=First LoRA filename on the volume"`
	LoraStrength1  float64 `json:"lora_strength_1" jsonschema:"required,description=First LoRA strength (0.0-1.0)"`
	LoraFilename2  string  `json:"lora_filename_2" jsonschema:"required,description=Second LoRA filename on the volume"`
	LoraStrength2  float64 `json:"lora_strength_2" jsonschema:"required,description=Second LoRA strength (0.0-1.0)"`
	LoraFilename3  string  `json:"lora_filename_3" jsonschema:"required,description=Third LoRA filename on the volume"`
	LoraStrength3  float64 `json:"lora_strength_3" jsonschema:"required,description=Third LoRA strength (0.0-1.0)"`
	OutputFilename string  `json:"output_filename" jsonschema:"description=Output filename. For repeats, variant suffix _vN will be appended (default: auto-generated)"`
}

type GenerateImageOutput struct {
	Message string `json:"message"`
}

type ListLorasOutput struct {
	Loras []LoraEntry `json:"loras"`
}
