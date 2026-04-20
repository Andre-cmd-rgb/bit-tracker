package ai

// ModelSpec is a curated entry in the download catalog.
//
// URLs point at public GGUF files on Hugging Face. Sizes are approximate.
type ModelSpec struct {
	ID       string // stable key, used in settings
	Name     string // display
	Params   string // parameter count label ("135M", "1.1B")
	Approx   string // on-disk size, human
	Filename string // final on-disk name
	URL      string // direct download
}

// Catalog is ordered from smallest to largest. All are quantised Q4_K_M.
func Catalog() []ModelSpec {
	return []ModelSpec{
		{
			ID:       "smollm2-135m",
			Name:     "SmolLM2 135M Instruct",
			Params:   "135M",
			Approx:   "~100 MB",
			Filename: "smollm2-135m-instruct-q4_k_m.gguf",
			URL:      "https://huggingface.co/HuggingFaceTB/SmolLM2-135M-Instruct-GGUF/resolve/main/smollm2-135m-instruct-q4_k_m.gguf?download=true",
		},
		{
			ID:       "smollm2-360m",
			Name:     "SmolLM2 360M Instruct",
			Params:   "360M",
			Approx:   "~230 MB",
			Filename: "smollm2-360m-instruct-q4_k_m.gguf",
			URL:      "https://huggingface.co/HuggingFaceTB/SmolLM2-360M-Instruct-GGUF/resolve/main/smollm2-360m-instruct-q4_k_m.gguf?download=true",
		},
		{
			ID:       "qwen2.5-0.5b",
			Name:     "Qwen2.5 0.5B Instruct",
			Params:   "500M",
			Approx:   "~400 MB",
			Filename: "qwen2.5-0.5b-instruct-q4_k_m.gguf",
			URL:      "https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct-GGUF/resolve/main/qwen2.5-0.5b-instruct-q4_k_m.gguf?download=true",
		},
		{
			ID:       "tinyllama-1.1b",
			Name:     "TinyLlama 1.1B Chat",
			Params:   "1.1B",
			Approx:   "~670 MB",
			Filename: "tinyllama-1.1b-chat-v1.0.q4_k_m.gguf",
			URL:      "https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf?download=true",
		},
		{
			ID:       "gemma-3-1b",
			Name:     "Gemma 3 1B Instruct",
			Params:   "1B",
			Approx:   "~800 MB",
			Filename: "gemma-3-1b-it-q4_k_m.gguf",
			URL:      "https://huggingface.co/unsloth/gemma-3-1b-it-GGUF/resolve/main/gemma-3-1b-it-Q4_K_M.gguf?download=true",
		},
		{
			ID:       "qwen2.5-1.5b",
			Name:     "Qwen2.5 1.5B Instruct",
			Params:   "1.5B",
			Approx:   "~1.0 GB",
			Filename: "qwen2.5-1.5b-instruct-q4_k_m.gguf",
			URL:      "https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf?download=true",
		},
	}
}

// FindModel returns the catalog entry by ID.
func FindModel(id string) (ModelSpec, bool) {
	for _, m := range Catalog() {
		if m.ID == id {
			return m, true
		}
	}
	return ModelSpec{}, false
}
