package ui

import "fmt"

// Logo prints the palm ASCII art banner with version info.
func Logo(version string, toolCount int) {
	tree := []string{
		`        _  _`,
		`      _/ \/ \_`,
		`     / \/ \/\ \`,
		`    /  /\  /\ \ \`,
		`   /_ /  \/  \_\_\`,
		`      \  ||  /`,
		`       \ || /`,
		`        \||/`,
		`         ||`,
		`         ||`,
		`         ||`,
		`        _||_`,
		`       |____|`,
	}

	palmText := []string{
		` ██████╗  █████╗ ██╗     ███╗   ███╗`,
		` ██╔══██╗██╔══██╗██║     ████╗ ████║`,
		` ██████╔╝███████║██║     ██╔████╔██║`,
		` ██╔═══╝ ██╔══██║██║     ██║╚██╔╝██║`,
		` ██║     ██║  ██║███████╗██║ ╚═╝ ██║`,
		` ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝`,
	}

	fmt.Println()
	for _, line := range tree {
		Brand.Println(line)
	}
	fmt.Println()
	for _, line := range palmText {
		Brand.Println(line)
	}
	fmt.Println()
	Subtle.Printf("  v%s", version)
	fmt.Printf("  %s  ", Palm)
	Subtle.Printf("%d tools in registry\n", toolCount)
	Subtle.Println("  The AI Tool Manager & Control Plane")
	fmt.Println()
}
