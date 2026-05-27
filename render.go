// render.go — output bonito do heatmap + stats no terminal.
// Usa lipgloss pra cores (true-color quando suportado, fallback ANSI 256).

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Paleta — mantém vibe do GitHub mas adapta pra terminal escuro.
// 5 níveis (0..4) matching o ContributionLevel.
var levelColors = []lipgloss.Color{
	lipgloss.Color("#21262d"), // 0 — vazio (cinza muito escuro, sutil mas visível)
	lipgloss.Color("#0e4429"), // 1 — quartil 1
	lipgloss.Color("#006d32"), // 2 — quartil 2
	lipgloss.Color("#26a641"), // 3 — quartil 3
	lipgloss.Color("#39d353"), // 4 — quartil 4 (verde mais vivo)
}

// blockChar é o caractere usado pra cada célula do heatmap.
// 2 chars wide pra ficar "quadradinho" no terminal monoespaçado.
const blockChar = "██"

// renderHeatmap monta o grid 7×~53 (dias da semana × semanas) colorido.
// Convenção GitHub: linhas = dom..sáb de cima pra baixo, colunas = semanas
// da mais antiga (esquerda) pra mais recente (direita).
func renderHeatmap(days []ContributionDay) string {
	if len(days) == 0 {
		return "(sem dados)"
	}

	// 1. Calcula o offset do primeiro dia (qual weekday é)
	// GitHub começa as semanas no domingo
	firstDay := days[0].Date
	startOffset := int(firstDay.Weekday()) // 0 = domingo

	// 2. Constrói matriz 7 × numWeeks
	totalDays := startOffset + len(days)
	numWeeks := (totalDays + 6) / 7

	// matriz[row][col] = level (-1 = vazio)
	matrix := make([][]int, 7)
	for i := range matrix {
		matrix[i] = make([]int, numWeeks)
		for j := range matrix[i] {
			matrix[i][j] = -1
		}
	}

	for i, d := range days {
		idx := startOffset + i
		row := idx % 7
		col := idx / 7
		matrix[row][col] = d.Level
	}

	// 3. Renderiza linha por linha
	var sb strings.Builder
	for row := 0; row < 7; row++ {
		// Label do dia da semana (alterna pra não poluir — só Seg/Qua/Sex)
		switch row {
		case 1:
			sb.WriteString("Seg ")
		case 3:
			sb.WriteString("Qua ")
		case 5:
			sb.WriteString("Sex ")
		default:
			sb.WriteString("    ")
		}

		for col := 0; col < numWeeks; col++ {
			level := matrix[row][col]
			if level == -1 {
				sb.WriteString("  ") // espaço-em-branco do mesmo tamanho do block
			} else {
				style := lipgloss.NewStyle().Foreground(levelColors[level])
				sb.WriteString(style.Render(blockChar))
			}
		}
		sb.WriteString("\n")
	}

	// 4. Label de meses no rodapé (simplificado — Jan, Fev, Mar...)
	sb.WriteString("    ")
	monthsRow := buildMonthsLabel(days, numWeeks, startOffset)
	sb.WriteString(monthsRow)
	sb.WriteString("\n")

	return sb.String()
}

// buildMonthsLabel monta a linha de meses alinhada com as colunas do heatmap.
// Cada célula tem 2 chars, e mostramos a abreviação do mês apenas na 1ª semana
// daquele mês.
func buildMonthsLabel(days []ContributionDay, numWeeks, startOffset int) string {
	monthNames := []string{
		"Jan", "Fev", "Mar", "Abr", "Mai", "Jun",
		"Jul", "Ago", "Set", "Out", "Nov", "Dez",
	}

	// Pra cada coluna, vê qual mês predomina (1ª dia útil daquela semana)
	colMonth := make([]int, numWeeks)
	for col := 0; col < numWeeks; col++ {
		idx := col*7 - startOffset
		if idx < 0 {
			idx = 0
		}
		if idx >= len(days) {
			break
		}
		colMonth[col] = int(days[idx].Date.Month()) - 1
	}

	var sb strings.Builder
	lastMonth := -1
	col := 0
	for col < numWeeks {
		m := colMonth[col]
		if m != lastMonth {
			// Nome do mês ocupa 4 chars (ex: "Jan ") = exatamente 2 cells
			sb.WriteString(monthNames[m] + " ")
			lastMonth = m
			col += 2
		} else {
			// 1 cell vazia
			sb.WriteString("  ")
			col++
		}
	}
	return sb.String()
}

// renderHeader mostra @login + nome + total contributions de forma bonita.
func renderHeader(data *ContributionData) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#39d353"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7d8590"))

	period := fmt.Sprintf("%s — %s",
		data.YearFrom.Format("Jan 02, 2006"),
		data.YearTo.Format("Jan 02, 2006"))

	name := data.Name
	if name == "" {
		name = data.Login
	}

	return fmt.Sprintf("%s %s\n%s · %s contributions\n\n",
		titleStyle.Render("@"+data.Login),
		dimStyle.Render("· "+name),
		dimStyle.Render(period),
		titleStyle.Render(formatInt(data.TotalCount)))
}

// renderStats mostra current + longest streak + porcentagem.
func renderStats(stats StreakStats) string {
	bold := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#39d353"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#7d8590"))
	fire := "🔥"

	var sb strings.Builder
	sb.WriteString("\n")

	// Current streak
	if stats.CurrentStreak > 0 {
		sb.WriteString(fmt.Sprintf("Current streak: %s %s\n",
			bold.Render(fmt.Sprintf("%d days", stats.CurrentStreak)),
			fire))
	} else {
		sb.WriteString(dim.Render("Current streak: 0 days — bora começar?\n"))
	}

	// Longest streak
	if stats.LongestStreak > 0 {
		period := fmt.Sprintf("%s → %s",
			stats.LongestStart.Format("Jan 2"),
			stats.LongestEnd.Format("Jan 2, 2006"))
		sb.WriteString(fmt.Sprintf("Longest streak: %s %s\n",
			bold.Render(fmt.Sprintf("%d days", stats.LongestStreak)),
			dim.Render("("+period+")")))
	}

	// Active days
	if stats.TotalDays > 0 {
		pct := float64(stats.ActiveDays) / float64(stats.TotalDays) * 100
		sb.WriteString(fmt.Sprintf("Active days:    %s %s\n",
			bold.Render(fmt.Sprintf("%d / %d", stats.ActiveDays, stats.TotalDays)),
			dim.Render(fmt.Sprintf("(%.0f%%)", pct))))
	}

	return sb.String()
}

// formatInt adiciona separador de milhar (1234567 → "1,234,567").
func formatInt(n int) string {
	if n < 0 {
		return "-" + formatInt(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}

// renderLegend mostra a barra "Less ▓▒░ More" no estilo GitHub.
func renderLegend() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#7d8590"))
	var sb strings.Builder
	sb.WriteString(dim.Render("Less "))
	for i := 0; i < 5; i++ {
		style := lipgloss.NewStyle().Foreground(levelColors[i])
		sb.WriteString(style.Render(blockChar))
	}
	sb.WriteString(dim.Render(" More"))
	return sb.String()
}

// timeNowYear retorna o ano atual (testável, mas mantemos simples por ora).
var timeNowYear = func() int { return time.Now().Year() }
