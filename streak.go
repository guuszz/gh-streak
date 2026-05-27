// streak.go — cálculo de streaks (current + longest) a partir dos days.

package main

import "time"

// StreakStats agrega métricas calculadas em cima do calendar.
type StreakStats struct {
	CurrentStreak int       // dias consecutivos até hoje (inclusivo)
	LongestStreak int       // maior sequência no período fetched
	LongestStart  time.Time // primeiro dia do longest streak
	LongestEnd    time.Time // último dia do longest streak
	ActiveDays    int       // dias com >= 1 contribution
	TotalDays     int       // dias no período
}

// Calculate computa as métricas em cima da lista ordenada de days.
//
// Regras:
//   - Current streak conta de hoje pra trás, parando no primeiro dia com 0
//     contributions. Hoje mesmo conta se já tiver commit; senão o streak
//     começa "ontem" (a regra do GitHub heatmap — tolerante até o fim do dia).
//   - Longest streak: maior run de dias consecutivos com count > 0.
//   - Days devem estar ordenados asc por data.
func Calculate(days []ContributionDay) StreakStats {
	stats := StreakStats{TotalDays: len(days)}
	if len(days) == 0 {
		return stats
	}

	// 1. Active days
	for _, d := range days {
		if d.Count > 0 {
			stats.ActiveDays++
		}
	}

	// 2. Longest streak — single pass
	var run int
	var runStart time.Time
	for _, d := range days {
		if d.Count > 0 {
			if run == 0 {
				runStart = d.Date
			}
			run++
			if run > stats.LongestStreak {
				stats.LongestStreak = run
				stats.LongestStart = runStart
				stats.LongestEnd = d.Date
			}
		} else {
			run = 0
		}
	}

	// 3. Current streak — de hoje pra trás
	// Vamos do último dia da lista. Se ele é "hoje" ou "ontem" e tem count > 0,
	// começa o counter. Se "hoje" não tem count, GitHub ainda mostra streak ativo
	// até a meia-noite (regra de gozo: até o fim do dia local).
	current := 0
	for i := len(days) - 1; i >= 0; i-- {
		d := days[i]
		// pula "hoje" se sem count — não quebra o streak ainda
		if i == len(days)-1 && d.Count == 0 && isToday(d.Date) {
			continue
		}
		if d.Count > 0 {
			current++
		} else {
			break
		}
	}
	stats.CurrentStreak = current

	return stats
}

func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}
