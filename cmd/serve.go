package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/msalah0e/palm/internal/gpu"
	"github.com/msalah0e/palm/internal/serve"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run and manage local LLM models",
		Long: `Orchestrate local LLM runtimes (ollama, llama.cpp, vllm) with
automatic GPU detection and model management.

  palm serve start             # Start default model
  palm serve start --model codellama  # Start specific model
  palm serve stop              # Stop running server
  palm serve status            # Show status
  palm serve models            # List downloadable models
  palm serve pull llama3.3     # Download a model`,
	}

	cmd.AddCommand(
		serveStartCmd(),
		serveStopCmd(),
		serveStatusCmd(),
		serveModelsCmd(),
		servePullCmd(),
	)

	return cmd
}

func serveStartCmd() *cobra.Command {
	var (
		model  string
		useGPU bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a local LLM server",
		Run: func(cmd *cobra.Command, args []string) {
			rt := serve.DetectRuntime()
			if rt == nil {
				ui.Bad.Println("  No LLM runtime found")
				fmt.Println()
				fmt.Println("  Install one:")
				fmt.Println("    palm install ollama     (recommended)")
				fmt.Println("    brew install llama.cpp")
				os.Exit(1)
			}

			ui.Banner("serve start")
			fmt.Printf("  Runtime:  %s\n", ui.Brand.Sprint(rt.String()))

			if model == "" {
				model = "llama3.3"
			}

			// Auto-detect GPU if not explicitly set
			if !cmd.Flags().Changed("gpu") {
				useGPU = gpu.HasGPU()
			}

			gpuStr := "CPU only"
			if useGPU {
				gpuStr = "GPU accelerated"
			}
			fmt.Printf("  Model:    %s\n", ui.Brand.Sprint(model))
			fmt.Printf("  GPU:      %s\n", gpuStr)
			fmt.Println()

			c := rt.Start(model, useGPU)
			if c == nil {
				ui.Bad.Println("  Runtime does not support starting")
				os.Exit(1)
			}

			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			fmt.Printf("  Starting %s with %s...\n\n", ui.Brand.Sprint(rt.Name), model)

			if err := c.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					os.Exit(exitErr.ExitCode())
				}
				ui.Bad.Printf("  Failed to start: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&model, "model", "m", "", "Model to serve (default: llama3.3)")
	cmd.Flags().BoolVar(&useGPU, "gpu", false, "Force GPU acceleration")
	return cmd
}

func serveStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running LLM server",
		Run: func(cmd *cobra.Command, args []string) {
			rt := serve.DetectRuntime()
			if rt == nil {
				fmt.Println("  No runtime detected")
				return
			}

			if rt.Name == "ollama" {
				c := exec.Command(rt.Path, "stop")
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				if err := c.Run(); err != nil {
					// ollama might not have a stop command; try killing the process
					fmt.Println("  Stopping ollama serve...")
					_ = exec.Command("pkill", "-f", "ollama serve").Run()
				}
			}

			ui.Good.Printf("  %s Server stopped\n", ui.StatusIcon(true))
		},
	}
}

func serveStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show running models and GPU info",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("serve status")

			rt := serve.DetectRuntime()
			if rt == nil {
				fmt.Println("  No LLM runtime installed")
				fmt.Println("  Install: palm install ollama")
				return
			}

			fmt.Printf("  Runtime: %s\n", ui.Brand.Sprint(rt.String()))

			if rt.IsRunning() {
				fmt.Printf("  Status:  %s running\n", ui.StatusIcon(true))
			} else {
				fmt.Printf("  Status:  %s not running\n", ui.Subtle.Sprint("-"))
			}

			// Show GPU info
			gpus := gpu.Detect()
			if len(gpus) > 0 {
				fmt.Printf("  GPU:     %s %s\n", gpus[0].Vendor, gpus[0].Model)
			} else {
				fmt.Printf("  GPU:     none detected (CPU only)\n")
			}

			// Show running models
			if rt.IsRunning() {
				fmt.Println()
				c := rt.ListModels()
				if c != nil {
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					_ = c.Run()
				}
			}
		},
	}
}

func serveModelsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "models",
		Short: "List popular downloadable models",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("local models")

			models := serve.PopularModels()

			headers := []string{"Model", "Params", "Size", "Min VRAM", "Category"}
			var rows [][]string

			for _, m := range models {
				rows = append(rows, []string{
					m.ID,
					m.Params,
					m.Size,
					fmt.Sprintf("%dMB", m.MinVRAM),
					m.Category,
				})
			}

			ui.Table(headers, rows)

			fmt.Println()
			fmt.Println("  Download: palm serve pull <model>")
			fmt.Println("  Run:      palm serve start --model <model>")
		},
	}
}

func servePullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull <model>",
		Short: "Download a model",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			model := args[0]

			rt := serve.DetectRuntime()
			if rt == nil {
				ui.Bad.Println("  No LLM runtime found. Install ollama first:")
				fmt.Println("    palm install ollama")
				os.Exit(1)
			}

			ui.Banner("pulling model")
			fmt.Printf("  Model:   %s\n", ui.Brand.Sprint(model))
			fmt.Printf("  Runtime: %s\n\n", rt.String())

			// Check if model is in our catalog
			known := false
			for _, m := range serve.PopularModels() {
				if m.ID == model || strings.HasPrefix(model, m.ID) {
					known = true
					break
				}
			}
			if !known {
				ui.Warn.Printf("  %s %q is not in palm's model catalog (may still work)\n\n", ui.WarnIcon(), model)
			}

			c := rt.Pull(model)
			if c == nil {
				ui.Bad.Printf("  %s does not support pulling models directly\n", rt.Name)
				os.Exit(1)
			}

			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				ui.Bad.Printf("  Failed to pull model: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("\n  %s Model %s ready\n", ui.StatusIcon(true), model)
			fmt.Printf("  Run: palm serve start --model %s\n", model)
		},
	}
}
