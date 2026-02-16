package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Brand colors
var (
	Brand  = color.New(color.FgHiGreen, color.Bold)
	Subtle = color.New(color.FgHiBlack)
	Warn   = color.New(color.FgYellow)
	Info   = color.New(color.FgCyan)
	Good   = color.New(color.FgGreen)
	Bad    = color.New(color.FgRed)
)

const Palm = "\U0001F334" // 🌴

// Banner prints the tamr banner for AI commands.
func Banner(subtitle string) {
	fmt.Printf("%s %s — %s\n\n", Palm, Brand.Sprint("tamr"), subtitle)
}

// Table prints a simple aligned table.
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	headerLine := "  "
	sepLine := "  "
	for i, h := range headers {
		headerLine += fmt.Sprintf("%-*s  ", widths[i], h)
		sepLine += strings.Repeat("\u2500", widths[i]) + "  "
	}
	Subtle.Println(headerLine)
	Subtle.Println(sepLine)

	// Print rows
	for _, row := range rows {
		line := "  "
		for i, cell := range row {
			if i < len(widths) {
				line += fmt.Sprintf("%-*s  ", widths[i], cell)
			}
		}
		fmt.Println(line)
	}
}

// StatusIcon returns a status icon string.
func StatusIcon(ok bool) string {
	if ok {
		return Good.Sprint("\u2713")
	}
	return Bad.Sprint("\u2717")
}

// WarnIcon returns a warning icon.
func WarnIcon() string {
	return Warn.Sprint("\u26A0")
}
