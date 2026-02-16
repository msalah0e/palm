package gpu

import (
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// Info holds detected GPU information.
type Info struct {
	Vendor  string // NVIDIA, AMD, Apple, Intel
	Model   string // e.g., "RTX 4090", "M3 Max"
	VRAM    string // e.g., "24GB", "unified 36GB"
	Driver  string // driver version
	Compute string // CUDA 12.3, Metal 3, ROCm 6.0
}

// Detect returns GPU information for the current system.
func Detect() []Info {
	switch runtime.GOOS {
	case "darwin":
		return detectMacOS()
	case "linux":
		return detectLinux()
	case "windows":
		return detectWindows()
	}
	return nil
}

func detectMacOS() []Info {
	// Check for Apple Silicon / Metal
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return nil
	}

	output := string(out)
	var gpus []Info

	// Parse chipset/model
	model := extractField(output, `Chipset Model:\s*(.+)`)
	if model == "" {
		model = extractField(output, `Chip:\s*(.+)`)
	}

	vram := extractField(output, `VRAM.*?:\s*(.+)`)
	if vram == "" {
		// Apple Silicon uses unified memory
		memOut, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			memStr := strings.TrimSpace(string(memOut))
			// Convert bytes to GB
			var memBytes int64
			if _, err := parseMemBytes(memStr, &memBytes); err == nil {
				vram = formatGB(memBytes) + " (unified)"
			}
		}
	}

	metal := extractField(output, `Metal.*?:\s*(.+)`)

	if model != "" {
		gpu := Info{
			Model:   model,
			VRAM:    vram,
			Compute: metal,
		}
		if strings.Contains(strings.ToLower(model), "apple") || strings.Contains(strings.ToLower(model), "m1") || strings.Contains(strings.ToLower(model), "m2") || strings.Contains(strings.ToLower(model), "m3") || strings.Contains(strings.ToLower(model), "m4") {
			gpu.Vendor = "Apple"
			if gpu.Compute == "" {
				gpu.Compute = "Metal"
			}
		} else if strings.Contains(strings.ToLower(model), "amd") || strings.Contains(strings.ToLower(model), "radeon") {
			gpu.Vendor = "AMD"
		} else if strings.Contains(strings.ToLower(model), "intel") {
			gpu.Vendor = "Intel"
		}
		gpus = append(gpus, gpu)
	}

	return gpus
}

func detectLinux() []Info {
	var gpus []Info

	// Check NVIDIA
	if out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,driver_version", "--format=csv,noheader,nounits").Output(); err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			parts := strings.Split(line, ", ")
			if len(parts) >= 3 {
				gpus = append(gpus, Info{
					Vendor: "NVIDIA",
					Model:  strings.TrimSpace(parts[0]),
					VRAM:   strings.TrimSpace(parts[1]) + " MiB",
					Driver: strings.TrimSpace(parts[2]),
				})
			}
		}
		// Detect CUDA version
		if cudaOut, err := exec.Command("nvidia-smi", "--query-gpu=compute_cap", "--format=csv,noheader").Output(); err == nil {
			cuda := strings.TrimSpace(string(cudaOut))
			for i := range gpus {
				gpus[i].Compute = "CUDA " + cuda
			}
		}
		return gpus
	}

	// Check AMD ROCm
	if out, err := exec.Command("rocm-smi", "--showproductname", "--showmeminfo", "vram").Output(); err == nil {
		output := string(out)
		model := extractField(output, `Card.*?:\s*(.+)`)
		gpus = append(gpus, Info{
			Vendor:  "AMD",
			Model:   model,
			Compute: "ROCm",
		})
		return gpus
	}

	// Check lspci as fallback
	if out, err := exec.Command("lspci").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "vga") || strings.Contains(lower, "3d") {
				gpus = append(gpus, Info{
					Model: strings.TrimSpace(line),
				})
			}
		}
	}

	return gpus
}

func detectWindows() []Info {
	var gpus []Info

	// Check NVIDIA
	if out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,driver_version", "--format=csv,noheader,nounits").Output(); err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			parts := strings.Split(line, ", ")
			if len(parts) >= 3 {
				gpus = append(gpus, Info{
					Vendor:  "NVIDIA",
					Model:   strings.TrimSpace(parts[0]),
					VRAM:    strings.TrimSpace(parts[1]) + " MiB",
					Driver:  strings.TrimSpace(parts[2]),
					Compute: "CUDA",
				})
			}
		}
		return gpus
	}

	// Fallback: use wmic
	if out, err := exec.Command("wmic", "path", "win32_VideoController", "get", "name,adapterram").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "AdapterRAM") {
				gpus = append(gpus, Info{
					Model:   line,
					Compute: "DirectML",
				})
			}
		}
	}

	return gpus
}

func extractField(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func parseMemBytes(s string, out *int64) (int, error) {
	var n int
	n, err := strings.NewReader(s).Read(make([]byte, len(s)))
	if err != nil {
		return 0, err
	}
	// Simple parse: atoi
	var val int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			val = val*10 + int64(c-'0')
		}
	}
	*out = val
	return n, nil
}

func formatGB(bytes int64) string {
	gb := bytes / (1024 * 1024 * 1024)
	if gb > 0 {
		return strings.TrimRight(strings.TrimRight(strings.Replace(
			strings.Replace(string(rune(gb/10+'0'))+string(rune(gb%10+'0')), "00", "0", 1),
			"0", "", -1), ""), ".") + "GB"
	}
	return "unknown"
}

// HasGPU returns true if any GPU was detected.
func HasGPU() bool {
	return len(Detect()) > 0
}

// RecommendModel suggests a model based on available VRAM.
func RecommendModel(vramMB int) string {
	switch {
	case vramMB >= 48000:
		return "llama3.3:70b"
	case vramMB >= 24000:
		return "llama3.3:70b-q4"
	case vramMB >= 16000:
		return "llama3.3"
	case vramMB >= 8000:
		return "llama3.2"
	case vramMB >= 4000:
		return "phi3:mini"
	default:
		return "tinyllama"
	}
}
