package top

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/msalah0e/palm/internal/gpu"
)

// ProcessInfo holds information about a running AI tool process.
type ProcessInfo struct {
	PID    int
	Name   string  // matched tool display name
	Binary string  // actual binary name
	CPU    float64 // CPU%
	Mem    float64 // MEM%
	MemMB  float64 // RSS in MB
	Cmd    string  // truncated command line
}

// SystemStats holds system resource usage.
type SystemStats struct {
	CPUPercent float64
	MemTotal   uint64
	MemUsed    uint64
	MemPercent float64
	CPUCores   int
	GPUs       []gpu.Info
}

// Config configures the top monitor.
type Config struct {
	RefreshInterval time.Duration
	KnownBinaries  map[string]string // binary name → display name
}

var (
	brand  = color.New(color.FgHiGreen, color.Bold)
	subtle = color.New(color.FgHiBlack)
	dim    = color.New(color.FgWhite)
	cyan   = color.New(color.FgCyan)
	yellow = color.New(color.FgYellow)
)

// Run starts the live top monitor loop.
func Run(cfg Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h\n")

	// Initial GPU detection (expensive, do once)
	gpus := gpu.Detect()

	ticker := time.NewTicker(cfg.RefreshInterval)
	defer ticker.Stop()

	// Render immediately, then on each tick
	render(scanProcesses(cfg.KnownBinaries), getSystemStats(gpus), cfg)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			render(scanProcesses(cfg.KnownBinaries), getSystemStats(gpus), cfg)
		}
	}
}

func scanProcesses(known map[string]string) []ProcessInfo {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"ps", "aux"}
	default:
		args = []string{"ps", "aux", "--no-headers"}
	}

	out, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var procs []ProcessInfo

	for i, line := range lines {
		// Skip header
		if i == 0 && strings.HasPrefix(line, "USER") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pid, _ := strconv.Atoi(fields[1])
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)
		rss, _ := strconv.ParseFloat(fields[5], 64)
		memMB := rss / 1024 // RSS is in KB

		// Full command is everything from field 10 onwards
		cmd := strings.Join(fields[10:], " ")

		// Extract the binary name from the command
		binary := extractBinary(cmd)

		// Check if this binary matches any known AI tool
		if displayName, ok := matchProcess(binary, cmd, known); ok {
			// Truncate command for display
			displayCmd := cmd
			if len(displayCmd) > 60 {
				displayCmd = displayCmd[:57] + "..."
			}

			procs = append(procs, ProcessInfo{
				PID:    pid,
				Name:   displayName,
				Binary: binary,
				CPU:    cpu,
				Mem:    mem,
				MemMB:  memMB,
				Cmd:    displayCmd,
			})
		}
	}

	// Sort by CPU descending
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].CPU > procs[j].CPU
	})

	return procs
}

func extractBinary(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	bin := parts[0]
	// Get just the filename
	if idx := strings.LastIndex(bin, "/"); idx >= 0 {
		bin = bin[idx+1:]
	}
	return bin
}

func matchProcess(binary, fullCmd string, known map[string]string) (string, bool) {
	// Direct binary match
	if name, ok := known[binary]; ok {
		return name, true
	}

	// For python/node wrappers, check the script/module name
	lowerBin := strings.ToLower(binary)
	if lowerBin == "python" || lowerBin == "python3" || lowerBin == "node" || lowerBin == "sh" || lowerBin == "bash" {
		// Check if any known binary appears in the full command
		lowerCmd := strings.ToLower(fullCmd)
		for knownBin, name := range known {
			if strings.Contains(lowerCmd, knownBin) {
				return name, true
			}
		}
	}

	return "", false
}

func getSystemStats(gpus []gpu.Info) SystemStats {
	stats := SystemStats{
		CPUCores: runtime.NumCPU(),
		GPUs:     gpus,
	}

	switch runtime.GOOS {
	case "darwin":
		stats.CPUPercent, stats.MemTotal, stats.MemUsed, stats.MemPercent = macOSStats()
	case "linux":
		stats.CPUPercent, stats.MemTotal, stats.MemUsed, stats.MemPercent = linuxStats()
	}

	return stats
}

func macOSStats() (cpuPct float64, memTotal, memUsed uint64, memPct float64) {
	// CPU from top -l 1
	if out, err := exec.Command("top", "-l", "1", "-n", "0", "-s", "0").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "CPU usage:") {
				// "CPU usage: 12.5% user, 8.3% sys, 79.1% idle"
				parts := strings.Fields(line)
				for i, p := range parts {
					if p == "idle" && i > 0 {
						idle, _ := strconv.ParseFloat(strings.TrimSuffix(parts[i-1], "%"), 64)
						cpuPct = 100.0 - idle
					}
				}
			}
		}
	}

	// Total memory from sysctl
	if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		bytes, _ := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
		memTotal = bytes / (1024 * 1024) // Convert to MB
	}

	// Used memory from vm_stat
	if out, err := exec.Command("vm_stat").Output(); err == nil {
		var active, wired, compressed uint64
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "Pages active:") {
				active = parseVMStatPages(line)
			} else if strings.Contains(line, "Pages wired") {
				wired = parseVMStatPages(line)
			} else if strings.Contains(line, "Pages occupied by compressor:") {
				compressed = parseVMStatPages(line)
			}
		}
		// Each page is 16384 bytes on Apple Silicon, 4096 on Intel
		pageSize := uint64(16384)
		if out, err := exec.Command("sysctl", "-n", "hw.pagesize").Output(); err == nil {
			if ps, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64); err == nil {
				pageSize = ps
			}
		}
		memUsed = (active + wired + compressed) * pageSize / (1024 * 1024)
	}

	if memTotal > 0 {
		memPct = float64(memUsed) / float64(memTotal) * 100
	}

	return
}

func parseVMStatPages(line string) uint64 {
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return 0
	}
	numStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(parts[1]), "."))
	val, _ := strconv.ParseUint(numStr, 10, 64)
	return val
}

func linuxStats() (cpuPct float64, memTotal, memUsed uint64, memPct float64) {
	// CPU from /proc/stat snapshot
	if out, err := os.ReadFile("/proc/stat"); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "cpu ") {
			fields := strings.Fields(lines[0])
			if len(fields) >= 8 {
				user, _ := strconv.ParseFloat(fields[1], 64)
				nice, _ := strconv.ParseFloat(fields[2], 64)
				system, _ := strconv.ParseFloat(fields[3], 64)
				idle, _ := strconv.ParseFloat(fields[4], 64)
				total := user + nice + system + idle
				if total > 0 {
					cpuPct = (total - idle) / total * 100
				}
			}
		}
	}

	// Memory from /proc/meminfo
	if out, err := os.ReadFile("/proc/meminfo"); err == nil {
		var total, available uint64
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "MemTotal:") {
				total = parseProcMemKB(line)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				available = parseProcMemKB(line)
			}
		}
		memTotal = total / 1024 // KB to MB
		if total > 0 {
			memUsed = (total - available) / 1024
			memPct = float64(total-available) / float64(total) * 100
		}
	}

	return
}

func parseProcMemKB(line string) uint64 {
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		val, _ := strconv.ParseUint(parts[1], 10, 64)
		return val
	}
	return 0
}

func render(procs []ProcessInfo, stats SystemStats, cfg Config) {
	// Move cursor to top-left and clear screen
	fmt.Print("\033[H\033[J")

	now := time.Now().Format("15:04:05")
	width := 66

	// Header
	brand.Printf("  \U0001F334 palm top \u2014 AI Tool Monitor")
	fmt.Printf("%*s\n", width-33, now)
	subtle.Println("  " + strings.Repeat("\u2500", width-2))

	// CPU bar
	cpuBar := progressBar(stats.CPUPercent, 20)
	fmt.Printf("  CPU  %s %5.1f%%  (%d cores)\n", cpuBar, stats.CPUPercent, stats.CPUCores)

	// Memory bar
	memBar := progressBar(stats.MemPercent, 20)
	totalGB := float64(stats.MemTotal) / 1024
	usedGB := float64(stats.MemUsed) / 1024
	fmt.Printf("  MEM  %s %5.1f%%  (%.1f / %.1f GB)\n", memBar, stats.MemPercent, usedGB, totalGB)

	// GPU line
	if len(stats.GPUs) > 0 {
		g := stats.GPUs[0]
		gpuLine := "  GPU  "
		if g.Model != "" {
			gpuLine += g.Model
		}
		if g.Compute != "" {
			gpuLine += " \u00b7 " + g.Compute
		}
		if g.VRAM != "" {
			gpuLine += " \u00b7 " + g.VRAM
		}
		cyan.Println(gpuLine)
	}

	subtle.Println("  " + strings.Repeat("\u2500", width-2))

	// Process table header
	if len(procs) > 0 {
		fmt.Printf("  %-7s %-18s %6s %7s %9s  %s\n",
			subtle.Sprint("PID"),
			subtle.Sprint("NAME"),
			subtle.Sprint("CPU%"),
			subtle.Sprint("MEM%"),
			subtle.Sprint("MEM(MB)"),
			subtle.Sprint("CMD"),
		)

		for _, p := range procs {
			cpuColor := dim
			if p.CPU > 50 {
				cpuColor = color.New(color.FgRed)
			} else if p.CPU > 20 {
				cpuColor = yellow
			}

			fmt.Printf("  %-7d %-18s %s %6.1f%% %8.0f  %s\n",
				p.PID,
				brand.Sprint(truncate(p.Name, 18)),
				cpuColor.Sprintf("%5.1f%%", p.CPU),
				p.Mem,
				p.MemMB,
				subtle.Sprint(p.Cmd),
			)
		}
	} else {
		subtle.Println("  No AI processes detected")
	}

	subtle.Println("  " + strings.Repeat("\u2500", width-2))

	// Footer
	interval := cfg.RefreshInterval.String()
	procCount := len(procs)
	subtle.Printf("  %d AI process", procCount)
	if procCount != 1 {
		subtle.Print("es")
	}
	subtle.Printf(" \u00b7 Refresh: %s \u00b7 Ctrl+C to exit\n", interval)
}

func progressBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := "[" + brand.Sprint(strings.Repeat("\u2588", filled)) +
		subtle.Sprint(strings.Repeat("\u2591", width-filled)) + "]"
	return bar
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "\u2026"
}
