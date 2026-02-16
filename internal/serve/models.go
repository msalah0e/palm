package serve

// Model represents a downloadable local LLM model.
type Model struct {
	ID       string
	Name     string
	Size     string // download size
	Params   string // parameter count
	Quant    string // quantization level
	MinVRAM  int    // minimum VRAM in MB
	Category string // chat, code, embedding
}

// PopularModels returns a curated list of recommended local models.
func PopularModels() []Model {
	return []Model{
		{ID: "llama3.3", Name: "Llama 3.3", Size: "4.7GB", Params: "8B", Quant: "Q4_0", MinVRAM: 8000, Category: "chat"},
		{ID: "llama3.3:70b", Name: "Llama 3.3 70B", Size: "39GB", Params: "70B", Quant: "Q4_0", MinVRAM: 48000, Category: "chat"},
		{ID: "codellama", Name: "Code Llama", Size: "3.8GB", Params: "7B", Quant: "Q4_0", MinVRAM: 6000, Category: "code"},
		{ID: "deepseek-coder-v2", Name: "DeepSeek Coder V2", Size: "8.9GB", Params: "16B", Quant: "Q4_0", MinVRAM: 12000, Category: "code"},
		{ID: "mistral", Name: "Mistral 7B", Size: "4.1GB", Params: "7B", Quant: "Q4_0", MinVRAM: 6000, Category: "chat"},
		{ID: "mixtral", Name: "Mixtral 8x7B", Size: "26GB", Params: "47B", Quant: "Q4_0", MinVRAM: 32000, Category: "chat"},
		{ID: "phi3:mini", Name: "Phi-3 Mini", Size: "2.3GB", Params: "3.8B", Quant: "Q4_0", MinVRAM: 4000, Category: "chat"},
		{ID: "qwen2.5-coder", Name: "Qwen 2.5 Coder", Size: "4.7GB", Params: "7B", Quant: "Q4_0", MinVRAM: 8000, Category: "code"},
		{ID: "nomic-embed-text", Name: "Nomic Embed", Size: "274MB", Params: "137M", Quant: "F16", MinVRAM: 1000, Category: "embedding"},
		{ID: "tinyllama", Name: "TinyLlama", Size: "637MB", Params: "1.1B", Quant: "Q4_0", MinVRAM: 2000, Category: "chat"},
	}
}
