package cmd

import (
	"fmt"
	"strings"

	"github.com/msalah0e/palm/internal/gpu"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func gpuCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gpu",
		Short: "Detect and display GPU information",
		Long: `Detect available GPUs and their capabilities for local LLM inference.

  palm gpu                 # Show GPU info and model recommendations`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("GPU detection")

			gpus := gpu.Detect()

			if len(gpus) == 0 {
				fmt.Println("  No GPU detected")
				fmt.Println()
				fmt.Println("  Local LLM inference will use CPU only.")
				fmt.Println("  For GPU acceleration, ensure drivers are installed.")
				return
			}

			headers := []string{"Vendor", "Model", "VRAM", "Compute"}
			var rows [][]string

			for _, g := range gpus {
				vendor := g.Vendor
				if vendor == "" {
					vendor = "Unknown"
				}
				vram := g.VRAM
				if vram == "" {
					vram = "-"
				}
				compute := g.Compute
				if compute == "" {
					compute = "-"
				}

				rows = append(rows, []string{vendor, g.Model, vram, compute})
			}

			ui.Table(headers, rows)

			// Show recommendation
			fmt.Println()
			fmt.Printf("  %s Recommended models for your hardware:\n", ui.Brand.Sprint("🌴"))

			g := gpus[0]
			if strings.Contains(strings.ToLower(g.Vendor), "apple") {
				fmt.Println("    - llama3.3       (8B, runs great on Apple Silicon)")
				fmt.Println("    - deepseek-coder (for coding tasks)")
				fmt.Println("    - phi3:mini      (lightweight, fast)")
			} else if strings.Contains(g.VRAM, "24") || strings.Contains(g.VRAM, "48") {
				fmt.Println("    - llama3.3:70b   (70B, needs 48GB+ VRAM)")
				fmt.Println("    - llama3.3       (8B, fast on your GPU)")
				fmt.Println("    - mixtral        (47B MoE, needs 32GB+)")
			} else {
				fmt.Println("    - llama3.3       (8B, good general model)")
				fmt.Println("    - phi3:mini      (3.8B, lightweight)")
				fmt.Println("    - tinyllama      (1.1B, minimal resources)")
			}

			fmt.Println()
			fmt.Println("  Install: palm serve pull <model>")
		},
	}
}
