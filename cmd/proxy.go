package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/msalah0e/palm/internal/proxy"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func proxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Local LLM API proxy — intercept, log, and budget API calls",
	}

	cmd.AddCommand(
		proxyStartCmd(),
		proxyStopCmd(),
		proxyStatusCmd(),
		proxyLogsCmd(),
	)

	return cmd
}

func proxyStartCmd() *cobra.Command {
	var port int
	var verbose bool
	var background bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the proxy server",
		Run: func(cmd *cobra.Command, args []string) {
			// Check if already running
			if running, pid := proxy.IsRunning(); running {
				fmt.Printf("  Proxy already running (PID %d)\n", pid)
				return
			}

			if background {
				// Launch in background
				exe, _ := os.Executable()
				child := exec.Command(exe, "proxy", "start", "--port", strconv.Itoa(port))
				if verbose {
					child.Args = append(child.Args, "--verbose")
				}
				child.Stdout = nil
				child.Stderr = nil
				child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

				if err := child.Start(); err != nil {
					ui.Bad.Printf("  Failed to start proxy: %v\n", err)
					os.Exit(1)
				}

				// Write PID
				_ = os.WriteFile(proxy.PidFile(), []byte(strconv.Itoa(child.Process.Pid)), 0o644)

				ui.Good.Printf("  %s Proxy started on port %d (PID %d)\n", ui.StatusIcon(true), port, child.Process.Pid)
				fmt.Println()
				fmt.Printf("  Set base URLs to route through proxy:\n")
				fmt.Printf("    export OPENAI_BASE_URL=http://localhost:%d/openai/v1\n", port)
				fmt.Printf("    export ANTHROPIC_BASE_URL=http://localhost:%d/anthropic/v1\n", port)
				return
			}

			// Foreground mode
			ui.Banner("proxy server")
			_ = proxy.WritePid()

			srv := proxy.New(proxy.Config{
				Port:    port,
				Verbose: verbose,
			})

			if err := srv.Start(); err != nil {
				ui.Bad.Printf("  Proxy error: %v\n", err)
				_ = os.Remove(proxy.PidFile())
				os.Exit(1)
			}
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 4778, "Port to listen on")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Log all requests to stdout")
	cmd.Flags().BoolVarP(&background, "bg", "b", false, "Run in background")
	return cmd
}

func proxyStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy server",
		Run: func(cmd *cobra.Command, args []string) {
			running, pid := proxy.IsRunning()
			if !running {
				fmt.Println("  Proxy is not running")
				return
			}

			proc, err := os.FindProcess(pid)
			if err != nil {
				ui.Bad.Printf("  Failed to find process %d: %v\n", pid, err)
				os.Exit(1)
			}

			if err := proc.Signal(syscall.SIGTERM); err != nil {
				ui.Bad.Printf("  Failed to stop proxy: %v\n", err)
				os.Exit(1)
			}

			_ = os.Remove(proxy.PidFile())
			ui.Good.Printf("  %s Proxy stopped (PID %d)\n", ui.StatusIcon(true), pid)
		},
	}
}

func proxyStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check proxy server status",
		Run: func(cmd *cobra.Command, args []string) {
			running, pid := proxy.IsRunning()
			if running {
				ui.Good.Printf("  %s Proxy running (PID %d)\n", ui.StatusIcon(true), pid)
				fmt.Println("  Routes: /openai/, /anthropic/, /google/, /groq/, /mistral/, /ollama/")
			} else {
				fmt.Println("  Proxy is not running")
				fmt.Println("  Start: palm proxy start")
			}
		},
	}
}

func proxyLogsCmd() *cobra.Command {
	var count int

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show recent proxy request logs",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("proxy logs")

			logs, err := proxy.ReadLogs(count)
			if err != nil {
				ui.Bad.Printf("  Failed to read logs: %v\n", err)
				os.Exit(1)
			}

			if len(logs) == 0 {
				fmt.Println("  No proxy logs yet.")
				return
			}

			headers := []string{"Time", "Provider", "Method", "Path", "Status", "Duration"}
			var rows [][]string

			for _, entry := range logs {
				statusIcon := ui.StatusIcon(entry.Status < 400)
				rows = append(rows, []string{
					entry.Timestamp.Format("15:04:05"),
					entry.Provider,
					entry.Method,
					truncate(entry.Path, 30),
					fmt.Sprintf("%s %d", statusIcon, entry.Status),
					fmt.Sprintf("%.0fms", entry.Duration),
				})
			}

			ui.Table(headers, rows)
			fmt.Printf("\n  %d entries\n", len(logs))
		},
	}

	cmd.Flags().IntVarP(&count, "count", "n", 50, "Number of log entries to show")
	return cmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
