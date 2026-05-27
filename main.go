// gh-streak — CLI que mostra teu streak de commits do GitHub no terminal.
//
// Uso:
//   gh-streak                          # tua own conta
//   gh-streak <login>                  # outro user
//   gh-streak --year 2025              # ano específico
//   gh-streak --no-color               # disable cores (pra piping)
//   gh-streak --json <login>           # output JSON pra script
//
// Token vem de (em ordem):
//   1. GH_TOKEN env var
//   2. GITHUB_TOKEN env var
//   3. `gh auth token` (GitHub CLI)

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var version = "dev" // sobrescrito via -ldflags em release

func main() {
	var (
		year     = flag.Int("year", 0, "specific year (default: last 365 days)")
		noColor  = flag.Bool("no-color", false, "disable colors")
		jsonOut  = flag.Bool("json", false, "output JSON instead of pretty render")
		showVer  = flag.Bool("version", false, "show version and exit")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `gh-streak %s — GitHub contribution streak no terminal

USAGE:
    gh-streak [FLAGS] [LOGIN]

ARGS:
    LOGIN    GitHub username (default: deduced from gh auth or git config)

FLAGS:
    --year=N      Specific year (default: last 365 days)
    --no-color    Disable ANSI colors (pra piping)
    --json        Output JSON
    --version     Show version
    --help        Show this

EXAMPLES:
    gh-streak                  # você mesmo
    gh-streak torvalds         # qualquer user
    gh-streak --year 2024      # ano específico
    gh-streak --json | jq      # script-friendly

TOKEN:
    Lê de GH_TOKEN, GITHUB_TOKEN, ou via "gh auth token".
    Scope mínimo: read:user

REPO: https://github.com/guuszz/gh-streak
`, version)
	}
	flag.Parse()

	if *showVer {
		fmt.Println("gh-streak", version)
		return
	}

	if *noColor {
		// lipgloss respeita a env var NO_COLOR automaticamente
		os.Setenv("NO_COLOR", "1")
	}

	login := flag.Arg(0)
	if login == "" {
		var err error
		login, err = deduceLogin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n\nPass a login as argument: gh-streak <username>\n", err)
			os.Exit(1)
		}
	}

	data, err := FetchContributions(login, *year)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	stats := Calculate(data.Days)

	if *jsonOut {
		printJSON(data, stats)
		return
	}

	// Output bonito
	fmt.Print(renderHeader(data))
	fmt.Print(renderHeatmap(data.Days))
	fmt.Println()
	fmt.Println(renderLegend())
	fmt.Print(renderStats(stats))
}

// deduceLogin tenta descobrir o login do user atual via gh CLI ou git config.
func deduceLogin() (string, error) {
	// 1. gh api user (preciso e oficial)
	if out, err := exec.Command("gh", "api", "user", "--jq", ".login").Output(); err == nil {
		login := strings.TrimSpace(string(out))
		if login != "" {
			return login, nil
		}
	}

	// 2. git config user.name como fallback (heurístico — usuário pode ter nome != login)
	if out, err := exec.Command("git", "config", "user.name").Output(); err == nil {
		name := strings.TrimSpace(string(out))
		// Só usa se for um único token sem espaço (login-like)
		if name != "" && !strings.Contains(name, " ") {
			return name, nil
		}
	}

	return "", fmt.Errorf("could not deduce login automatically")
}

// printJSON emite um payload script-friendly em stdout.
func printJSON(data *ContributionData, stats StreakStats) {
	out := map[string]any{
		"login":       data.Login,
		"name":        data.Name,
		"total":       data.TotalCount,
		"period_from": data.YearFrom.Format("2006-01-02"),
		"period_to":   data.YearTo.Format("2006-01-02"),
		"streak": map[string]any{
			"current":         stats.CurrentStreak,
			"longest":         stats.LongestStreak,
			"longest_start":   stats.LongestStart.Format("2006-01-02"),
			"longest_end":     stats.LongestEnd.Format("2006-01-02"),
			"active_days":     stats.ActiveDays,
			"total_days":      stats.TotalDays,
			"active_pct":      pct(stats.ActiveDays, stats.TotalDays),
		},
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding JSON: %s\n", err)
		os.Exit(1)
	}
}

func pct(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b) * 100
}
