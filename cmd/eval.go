package cmd

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

func evalCmd() *cobra.Command {
	var (
		tools   string
		context string
		judge   string
		timeout int
	)

	cmd := &cobra.Command{
		Use:   `eval "<question>" --tools tool1,tool2`,
		Short: "Evaluate AI accuracy — hallucination detection and quality scoring",
		Long: `Evaluate AI tool outputs for accuracy, hallucination, and quality.
Sends the same question to multiple AI tools, then uses a judge to score each response
on factual accuracy, hallucination level, completeness, and clarity.

This gives you a trustworthiness score for each tool on your specific use cases.

Examples:
  palm eval "What is the capital of France?" --tools ollama,mods
  palm eval "Explain how TCP works" --tools ollama,aider --context "networking basics"
  palm eval "What year was Python released?" --tools ollama,mods --judge ollama`,
		Aliases: []string{"evaluate", "check"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			question := args[0]
			toolNames := strings.Split(tools, ",")

			if len(toolNames) < 1 {
				ui.Warn.Println("  Provide at least 1 tool: --tools tool1[,tool2]")
				os.Exit(1)
			}

			for i := range toolNames {
				toolNames[i] = strings.TrimSpace(toolNames[i])
			}

			if judge == "" {
				judge = toolNames[0] // Use first tool as judge if not specified
			}

			ui.Banner("eval")
			printEvalHeader()
			fmt.Println()
			fmt.Printf("  Question: %s\n", ui.Brand.Sprint(question))
			if context != "" {
				fmt.Printf("  Context:  %s\n", ui.Subtle.Sprint(context))
			}
			fmt.Printf("  Tools:    %s\n", strings.Join(toolNames, ", "))
			fmt.Printf("  Judge:    %s\n", ui.Info.Sprint(judge))
			fmt.Println()

			reg := loadRegistry()
			v := vault.New()
			env := buildVaultEnv(v)

			// Run all tools on the question
			results := runSquad(toolNames, question, reg, env, timeout)

			// Now evaluate each result
			fmt.Println()
			fmt.Printf("  %s Evaluating responses...\n", ui.Info.Sprint("🔍"))
			fmt.Println()

			var scores []evalScore

			for _, r := range results {
				if r.Error != "" {
					scores = append(scores, evalScore{
						Tool:    r.Tool,
						Verdict: "FAILED: " + r.Error,
					})
					continue
				}

				// Build evaluation prompt
				evalPrompt := buildEvalPrompt(question, context, r.Output)
				judgeOutput := runJudgeTool(judge, evalPrompt, env, timeout)

				score := parseEvalScore(r.Tool, judgeOutput)
				scores = append(scores, score)
			}

			// Print scorecard
			printEvalScorecard(scores)
		},
	}

	cmd.Flags().StringVar(&tools, "tools", "", "Comma-separated list of tools to evaluate (required)")
	cmd.Flags().StringVar(&context, "context", "", "Additional context for evaluation")
	cmd.Flags().StringVar(&judge, "judge", "", "Tool to use as evaluator (default: first tool)")
	cmd.Flags().IntVar(&timeout, "timeout", 60, "Timeout per tool in seconds")
	_ = cmd.MarkFlagRequired("tools")
	return cmd
}

func printEvalHeader() {
	fmt.Println(ui.Brand.Sprint("  ╔═══════════════════════════════════════════════════╗"))
	fmt.Println(ui.Brand.Sprint("  ║") + "   🔬  " + ui.Brand.Sprint("palm eval") + " — AI Accuracy & Trust Scanner     " + ui.Brand.Sprint("║"))
	fmt.Println(ui.Brand.Sprint("  ╚═══════════════════════════════════════════════════╝"))
}

func buildEvalPrompt(question, context, response string) string {
	contextPart := ""
	if context != "" {
		contextPart = fmt.Sprintf("\nContext: %s\n", context)
	}

	return fmt.Sprintf(`You are an AI output evaluator. Score the following AI response on these criteria.

Question: "%s"%s

AI Response:
---
%s
---

Score each criterion from 0 to 100. Reply in EXACTLY this format (just the numbers and verdict, nothing else):

ACCURACY: [0-100]
HALLUCINATION: [0-100]
COMPLETENESS: [0-100]
CLARITY: [0-100]
VERDICT: [one sentence summary]

Scoring guide:
- ACCURACY: How factually correct is the response? 100 = perfectly accurate
- HALLUCINATION: How much fabricated/false info? 0 = no hallucination, 100 = entirely made up
- COMPLETENESS: Does it fully answer the question? 100 = thorough answer
- CLARITY: How clear and well-structured? 100 = crystal clear`, question, contextPart, response)
}

func runJudgeTool(judge, prompt string, env []string, timeout int) string {
	var cmdArgs []string
	switch judge {
	case "ollama":
		cmdArgs = []string{"ollama", "run", "llama3.3", prompt}
	default:
		cmdArgs = []string{judge, prompt}
	}

	var stdout bytes.Buffer
	c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	c.Stdout = &stdout
	c.Stderr = &bytes.Buffer{}
	c.Env = env
	c.Stdin = strings.NewReader(prompt)

	if err := c.Start(); err != nil {
		return ""
	}

	done := make(chan error, 1)
	go func() { done <- c.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			return ""
		}
		return stdout.String()
	case <-time.After(time.Duration(timeout) * time.Second):
		_ = c.Process.Kill()
		return ""
	}
}

type evalScore struct {
	Tool          string
	Accuracy      int
	Hallucination int
	Completeness  int
	Clarity       int
	Overall       int
	Verdict       string
}

func parseEvalScore(tool, output string) evalScore {
	score := evalScore{Tool: tool}

	// Parse scores from judge output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ACCURACY:") {
			fmt.Sscanf(strings.TrimPrefix(line, "ACCURACY:"), "%d", &score.Accuracy)
		} else if strings.HasPrefix(line, "HALLUCINATION:") {
			fmt.Sscanf(strings.TrimPrefix(line, "HALLUCINATION:"), "%d", &score.Hallucination)
		} else if strings.HasPrefix(line, "COMPLETENESS:") {
			fmt.Sscanf(strings.TrimPrefix(line, "COMPLETENESS:"), "%d", &score.Completeness)
		} else if strings.HasPrefix(line, "CLARITY:") {
			fmt.Sscanf(strings.TrimPrefix(line, "CLARITY:"), "%d", &score.Clarity)
		} else if strings.HasPrefix(line, "VERDICT:") {
			score.Verdict = strings.TrimSpace(strings.TrimPrefix(line, "VERDICT:"))
		}
	}

	// If parsing failed, use defaults based on whether output existed
	if score.Accuracy == 0 && score.Hallucination == 0 && output != "" {
		score.Accuracy = 50
		score.Hallucination = 50
		score.Completeness = 50
		score.Clarity = 50
		score.Verdict = "Could not parse judge output — showing estimates"
	}

	// Calculate overall (accuracy + completeness + clarity weighted, hallucination penalty)
	if score.Accuracy > 0 || score.Completeness > 0 {
		hallPenalty := float64(score.Hallucination) * 0.5
		raw := (float64(score.Accuracy)*0.4 + float64(score.Completeness)*0.3 + float64(score.Clarity)*0.3) - hallPenalty
		score.Overall = int(math.Max(0, math.Min(100, raw)))
	}

	return score
}

func printEvalScorecard(scores []evalScore) {
	fmt.Println(ui.Brand.Sprint("  ┌──────────────────────────────────────────────────────────────────┐"))
	fmt.Println(ui.Brand.Sprint("  │") + "  " + ui.Brand.Sprint("EVALUATION SCORECARD") + "                                            " + ui.Brand.Sprint("│"))
	fmt.Println(ui.Brand.Sprint("  ├──────────────────────────────────────────────────────────────────┤"))

	for _, s := range scores {
		fmt.Println(ui.Brand.Sprint("  │") + "                                                                  " + ui.Brand.Sprint("│"))

		if s.Accuracy == 0 && s.Hallucination == 0 && s.Verdict != "" && strings.HasPrefix(s.Verdict, "FAILED") {
			line := fmt.Sprintf("  %-14s  %s", s.Tool, ui.Bad.Sprint(s.Verdict))
			pad := max(0, 64-len(s.Tool)-2-len(s.Verdict)-2)
			fmt.Println(ui.Brand.Sprint("  │") + line + strings.Repeat(" ", pad) + ui.Brand.Sprint("│"))
			continue
		}

		// Tool name
		toolLine := fmt.Sprintf("  %s", ui.Brand.Sprint(s.Tool))
		overallGrade := gradeFromScore(s.Overall)
		gradePad := max(0, 60-len(s.Tool)-len(overallGrade)-4)
		fmt.Println(ui.Brand.Sprint("  │") + toolLine + strings.Repeat(" ", gradePad) + overallGrade + "  " + ui.Brand.Sprint("│"))

		// Score bars
		printScoreBar("  Accuracy", s.Accuracy, true)
		printScoreBar("  Hallucination", s.Hallucination, false) // Lower is better
		printScoreBar("  Completeness", s.Completeness, true)
		printScoreBar("  Clarity", s.Clarity, true)

		// Overall
		fmt.Printf(ui.Brand.Sprint("  │")+"  Overall: %s  %s\n",
			formatScoreNum(s.Overall),
			strings.Repeat(" ", max(0, 46))+ui.Brand.Sprint("│"))

		// Verdict
		if s.Verdict != "" {
			verdictLine := fmt.Sprintf("  💬 %s", s.Verdict)
			if len(verdictLine) > 62 {
				verdictLine = verdictLine[:62] + "..."
			}
			pad := max(0, 66-len(verdictLine))
			fmt.Println(ui.Brand.Sprint("  │") + verdictLine + strings.Repeat(" ", pad) + ui.Brand.Sprint("│"))
		}
	}

	fmt.Println(ui.Brand.Sprint("  │") + "                                                                  " + ui.Brand.Sprint("│"))
	fmt.Println(ui.Brand.Sprint("  └──────────────────────────────────────────────────────────────────┘"))

	// Trust recommendation
	fmt.Println()
	var bestTool string
	var bestScore int
	for _, s := range scores {
		if s.Overall > bestScore {
			bestScore = s.Overall
			bestTool = s.Tool
		}
	}
	if bestTool != "" {
		fmt.Printf("  %s Most trusted: %s (score: %d/100)\n",
			ui.Brand.Sprint("🏆"), ui.Brand.Sprint(bestTool), bestScore)
	}
	fmt.Println()
}

func printScoreBar(label string, score int, higherIsBetter bool) {
	barWidth := 20
	filled := score * barWidth / 100

	var bar string
	if higherIsBetter {
		if score >= 70 {
			bar = ui.Good.Sprint(strings.Repeat("█", filled))
		} else if score >= 40 {
			bar = ui.Warn.Sprint(strings.Repeat("█", filled))
		} else {
			bar = ui.Bad.Sprint(strings.Repeat("█", filled))
		}
	} else {
		// For hallucination, lower is better
		if score <= 20 {
			bar = ui.Good.Sprint(strings.Repeat("█", filled))
		} else if score <= 50 {
			bar = ui.Warn.Sprint(strings.Repeat("█", filled))
		} else {
			bar = ui.Bad.Sprint(strings.Repeat("█", filled))
		}
	}

	empty := ui.Subtle.Sprint(strings.Repeat("░", barWidth-filled))
	scoreStr := formatScoreNum(score)

	line := fmt.Sprintf("  %-16s %s%s %s", label, bar, empty, scoreStr)
	// Approximate padding — the ANSI codes make exact calc tricky
	fmt.Printf(ui.Brand.Sprint("  │")+"%s\n", line)
}

func formatScoreNum(score int) string {
	if score >= 70 {
		return ui.Good.Sprintf("%3d", score)
	} else if score >= 40 {
		return ui.Warn.Sprintf("%3d", score)
	}
	return ui.Bad.Sprintf("%3d", score)
}

func gradeFromScore(score int) string {
	switch {
	case score >= 90:
		return ui.Good.Sprint("A+")
	case score >= 80:
		return ui.Good.Sprint("A")
	case score >= 70:
		return ui.Good.Sprint("B")
	case score >= 60:
		return ui.Warn.Sprint("C")
	case score >= 50:
		return ui.Warn.Sprint("D")
	default:
		return ui.Bad.Sprint("F")
	}
}
