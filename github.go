// github.go — GraphQL client pra contributions calendar do GitHub.
// A REST API NÃO expõe o calendar — só GraphQL via /graphql.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const githubGraphQLEndpoint = "https://api.github.com/graphql"

// ContributionDay representa 1 dia no calendar.
type ContributionDay struct {
	Date  time.Time
	Count int
	// Color level conforme GitHub: NONE, FIRST_QUARTILE, SECOND_QUARTILE,
	// THIRD_QUARTILE, FOURTH_QUARTILE. Mapeado pra inteiro 0..4 pra render.
	Level int
}

// ContributionData é o resultado da query GraphQL.
type ContributionData struct {
	Login       string
	Name        string
	TotalCount  int
	Days        []ContributionDay // sorted asc por data
	YearFrom    time.Time
	YearTo      time.Time
}

// graphqlResponse mapeia o shape do GitHub GraphQL.
type graphqlResponse struct {
	Data struct {
		User struct {
			Login string `json:"login"`
			Name  string `json:"name"`
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []struct {
							ContributionCount int    `json:"contributionCount"`
							Date              string `json:"date"`
							ContributionLevel string `json:"contributionLevel"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// FetchContributions busca o contribution calendar do user.
// Se `year` for 0, pega os últimos 365 dias. Senão pega o ano calendário inteiro.
func FetchContributions(login string, year int) (*ContributionData, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	var fromTime, toTime time.Time
	if year == 0 {
		toTime = time.Now()
		fromTime = toTime.AddDate(-1, 0, 0)
	} else {
		fromTime = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		toTime = time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	query := fmt.Sprintf(`
query($login: String!, $from: DateTime!, $to: DateTime!) {
  user(login: $login) {
    login
    name
    contributionsCollection(from: $from, to: $to) {
      contributionCalendar {
        totalContributions
        weeks {
          contributionDays {
            contributionCount
            date
            contributionLevel
          }
        }
      }
    }
  }
}`)

	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"login": login,
			"from":  fromTime.Format(time.RFC3339),
			"to":    toTime.Format(time.RFC3339),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", githubGraphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gh-streak (https://github.com/guuszz/gh-streak)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed graphqlResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(parsed.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", parsed.Errors[0].Message)
	}

	if parsed.Data.User.Login == "" {
		return nil, fmt.Errorf("user %q not found", login)
	}

	// Achata weeks → days e converte tipos
	cal := parsed.Data.User.ContributionsCollection.ContributionCalendar
	days := make([]ContributionDay, 0, 366)
	for _, week := range cal.Weeks {
		for _, d := range week.ContributionDays {
			date, err := time.Parse("2006-01-02", d.Date)
			if err != nil {
				return nil, fmt.Errorf("parse date %q: %w", d.Date, err)
			}
			days = append(days, ContributionDay{
				Date:  date,
				Count: d.ContributionCount,
				Level: levelToInt(d.ContributionLevel),
			})
		}
	}

	return &ContributionData{
		Login:      parsed.Data.User.Login,
		Name:       parsed.Data.User.Name,
		TotalCount: cal.TotalContributions,
		Days:       days,
		YearFrom:   fromTime,
		YearTo:     toTime,
	}, nil
}

// levelToInt converte o enum ContributionLevel do GitHub pra 0..4.
func levelToInt(s string) int {
	switch s {
	case "NONE":
		return 0
	case "FIRST_QUARTILE":
		return 1
	case "SECOND_QUARTILE":
		return 2
	case "THIRD_QUARTILE":
		return 3
	case "FOURTH_QUARTILE":
		return 4
	}
	return 0
}

// getToken pega o token do GitHub via 3 vias, em ordem:
//   1. Env var GH_TOKEN ou GITHUB_TOKEN
//   2. `gh auth token` (GitHub CLI já autenticado)
//   3. Erro
//
// O scope mínimo necessário é `read:user` (pro contribution graph). `repo`
// também funciona mas é overkill.
func getToken() (string, error) {
	if t := os.Getenv("GH_TOKEN"); t != "" {
		return strings.TrimSpace(t), nil
	}
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return strings.TrimSpace(t), nil
	}
	// Tenta `gh auth token` — silencioso se gh não tá instalado
	out, err := exec.Command("gh", "auth", "token").Output()
	if err == nil {
		t := strings.TrimSpace(string(out))
		if t != "" {
			return t, nil
		}
	}
	return "", fmt.Errorf("no GitHub token found\n\nSetup options:\n  1. export GH_TOKEN=ghp_xxx (create at https://github.com/settings/tokens)\n  2. gh auth login (then re-run gh-streak)\n\nMinimum scope required: read:user")
}
