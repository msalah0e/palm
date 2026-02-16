package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/msalah0e/palm/internal/graph"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func graphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "graph",
		Short:   "Encrypted knowledge graph for AI memory",
		Aliases: []string{"kg", "memory"},
		Run: func(cmd *cobra.Command, args []string) {
			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			stats := g.GetStats()
			ui.Banner("knowledge graph")

			if stats.Entities == 0 {
				fmt.Println("  Empty graph. Get started:")
				fmt.Println()
				ui.Info.Println("  palm graph add <name> --type <type>")
				ui.Info.Println("  palm graph observe <name> \"fact\"")
				ui.Info.Println("  palm graph relate <from> <relation> <to>")
				return
			}

			fmt.Printf("  %s  %d\n", ui.Brand.Sprintf("%-16s", "Entities"), stats.Entities)
			fmt.Printf("  %s  %d\n", ui.Brand.Sprintf("%-16s", "Relations"), stats.Relations)
			fmt.Printf("  %s  %d\n", ui.Brand.Sprintf("%-16s", "Observations"), stats.Observations)
			fmt.Printf("  %s  %d\n", ui.Brand.Sprintf("%-16s", "Types"), stats.Types)
			fmt.Println()
			fmt.Printf("  %s\n", ui.Subtle.Sprint("Stored encrypted at ~/.config/palm/graph.enc"))
		},
	}

	cmd.AddCommand(
		graphAddCmd(),
		graphObserveCmd(),
		graphRelateCmd(),
		graphShowCmd(),
		graphSearchCmd(),
		graphListCmd(),
		graphRemoveCmd(),
		graphExportCmd(),
		graphImportCmd(),
		graphViewCmd(),
	)

	return cmd
}

func graphAddCmd() *cobra.Command {
	var entityType string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new entity",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if entityType == "" {
				entityType = "default"
			}

			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			if err := g.AddEntity(name, entityType); err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}

			if err := graph.Save(g); err != nil {
				ui.Bad.Printf("  Failed to save graph: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s Added %s (%s)\n", ui.StatusIcon(true), ui.Brand.Sprint(name), entityType)
		},
	}

	cmd.Flags().StringVar(&entityType, "type", "", "Entity type (e.g., person, project, tool)")
	return cmd
}

func graphObserveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "observe <name> <observation>",
		Short:   "Add an observation to an entity",
		Aliases: []string{"obs", "note"},
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name, observation := args[0], args[1]

			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			if err := g.AddObservation(name, observation); err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}

			if err := graph.Save(g); err != nil {
				ui.Bad.Printf("  Failed to save graph: %v\n", err)
				os.Exit(1)
			}

			e, _ := g.GetEntity(name)
			ui.Good.Printf("  %s Added observation to %s (%d total)\n", ui.StatusIcon(true), ui.Brand.Sprint(e.Name), len(e.Observations))
		},
	}
}

func graphRelateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "relate <from> <relation> <to>",
		Short: "Create a directed relation between entities",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			from, relType, to := args[0], args[1], args[2]

			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			if err := g.AddRelation(from, relType, to); err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}

			if err := graph.Save(g); err != nil {
				ui.Bad.Printf("  Failed to save graph: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s %s --%s--> %s\n", ui.StatusIcon(true), ui.Brand.Sprint(from), relType, ui.Brand.Sprint(to))
		},
	}
}

func graphShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show entity details and connections",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			if jsonOutput {
				result, err := g.ShowEntity(name)
				if err != nil {
					ui.Bad.Printf("  %v\n", err)
					os.Exit(1)
				}
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				return
			}

			output, err := graph.RenderShow(g, name,
				func(s string) string { return ui.Brand.Sprint(s) },
				func(s string) string { return ui.Subtle.Sprint(s) },
				func(s string) string { return ui.Info.Sprint(s) },
			)
			if err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}

			fmt.Println()
			fmt.Print(output)
			fmt.Println()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (for AI tools)")
	return cmd
}

func graphSearchCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search entities by name, type, or observation",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := args[0]

			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			results := g.Search(query)

			if jsonOutput {
				data, _ := json.MarshalIndent(results, "", "  ")
				fmt.Println(string(data))
				return
			}

			if len(results) == 0 {
				fmt.Printf("  No entities found matching %q\n", query)
				return
			}

			ui.Banner("search results")
			var rows [][]string
			for _, r := range results {
				obs := ""
				if len(r.Entity.Observations) > 0 {
					obs = r.Entity.Observations[0]
					if len(obs) > 40 {
						obs = obs[:37] + "..."
					}
				}
				rows = append(rows, []string{r.Entity.Name, r.Entity.Type, obs, fmt.Sprintf("%d", r.Score)})
			}
			ui.Table([]string{"Name", "Type", "Observation", "Score"}, rows)
			fmt.Printf("\n  %d results\n", len(results))
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (for AI tools)")
	return cmd
}

func graphListCmd() *cobra.Command {
	var filterType string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all entities",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			// Collect entities, optionally filtered
			var entities []*graph.Entity
			for _, name := range g.EntityNames() {
				e, _ := g.GetEntity(name)
				if filterType != "" && !strings.EqualFold(e.Type, filterType) {
					continue
				}
				entities = append(entities, e)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(entities, "", "  ")
				fmt.Println(string(data))
				return
			}

			if len(entities) == 0 {
				if filterType != "" {
					fmt.Printf("  No entities of type %q\n", filterType)
				} else {
					fmt.Println("  No entities in graph")
				}
				return
			}

			ui.Banner("entities")
			var rows [][]string
			for _, e := range entities {
				obs := fmt.Sprintf("%d", len(e.Observations))
				outgoing, incoming := g.RelationsOf(e.Name)
				rels := fmt.Sprintf("%d out / %d in", len(outgoing), len(incoming))
				rows = append(rows, []string{e.Name, e.Type, obs, rels})
			}
			ui.Table([]string{"Name", "Type", "Observations", "Relations"}, rows)
			fmt.Printf("\n  %d entities\n", len(entities))
		},
	}

	cmd.Flags().StringVar(&filterType, "type", "", "Filter by entity type")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (for AI tools)")
	return cmd
}

func graphRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Short:   "Remove an entity and its relations",
		Aliases: []string{"rm", "delete"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			if err := g.RemoveEntity(name); err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}

			if err := graph.Save(g); err != nil {
				ui.Bad.Printf("  Failed to save graph: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s Removed %s and its relations\n", ui.StatusIcon(true), name)
		},
	}
}

func graphExportCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the full graph (decrypted)",
		Run: func(cmd *cobra.Command, args []string) {
			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			switch format {
			case "json":
				data, err := g.ExportJSON()
				if err != nil {
					ui.Bad.Printf("  Export failed: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(string(data))
			case "dot":
				fmt.Print(g.ExportDOT())
			case "html":
				fmt.Print(g.ExportHTML())
			default:
				ui.Bad.Printf("  Unknown format: %s (use json, dot, or html)\n", format)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "Export format: json, dot, or html")
	return cmd
}

func graphImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <file>",
		Short: "Import/merge entities from a JSON file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filePath := args[0]

			data, err := os.ReadFile(filePath)
			if err != nil {
				ui.Bad.Printf("  Failed to read file: %v\n", err)
				os.Exit(1)
			}

			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			added, merged, relAdded, err := g.ImportJSON(data)
			if err != nil {
				ui.Bad.Printf("  Import failed: %v\n", err)
				os.Exit(1)
			}

			if err := graph.Save(g); err != nil {
				ui.Bad.Printf("  Failed to save graph: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s Imported: %d added, %d merged, %d relations\n",
				ui.StatusIcon(true), added, merged, relAdded)
		},
	}
}

func graphViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Open interactive graph visualization in browser (Obsidian-like)",
		Run: func(cmd *cobra.Command, args []string) {
			g, err := graph.Load()
			if err != nil {
				ui.Bad.Printf("  Failed to load graph: %v\n", err)
				os.Exit(1)
			}

			stats := g.GetStats()
			if stats.Entities == 0 {
				fmt.Println("  Empty graph — add some entities first")
				return
			}

			// Write HTML to temp file and open in browser
			tmpDir := os.TempDir()
			htmlPath := filepath.Join(tmpDir, "palm-graph.html")
			if err := os.WriteFile(htmlPath, []byte(g.ExportHTML()), 0o644); err != nil {
				ui.Bad.Printf("  Failed to write HTML: %v\n", err)
				os.Exit(1)
			}

			var openCmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				openCmd = exec.Command("open", htmlPath)
			case "linux":
				openCmd = exec.Command("xdg-open", htmlPath)
			default:
				// Windows or other
				openCmd = exec.Command("cmd", "/c", "start", htmlPath)
			}

			if err := openCmd.Start(); err != nil {
				// Fallback: just print the path
				fmt.Printf("  HTML written to: %s\n", htmlPath)
				fmt.Println("  Open it in your browser to see the graph")
				return
			}

			ui.Good.Printf("  %s Opened graph visualization (%d entities, %d relations)\n",
				ui.StatusIcon(true), stats.Entities, stats.Relations)
			ui.Subtle.Printf("  %s\n", htmlPath)
		},
	}
}
