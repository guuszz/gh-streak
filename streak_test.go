// streak_test.go — testes da lógica de cálculo de streak.

package main

import (
	"testing"
	"time"
)

// mkDay cria um ContributionDay relativo a hoje (daysAgo=0 é hoje).
func mkDay(daysAgo, count int) ContributionDay {
	return ContributionDay{
		Date:  time.Now().AddDate(0, 0, -daysAgo),
		Count: count,
	}
}

func TestCalculate_Empty(t *testing.T) {
	stats := Calculate(nil)
	if stats.TotalDays != 0 || stats.CurrentStreak != 0 || stats.LongestStreak != 0 || stats.ActiveDays != 0 {
		t.Fatalf("lista vazia deveria zerar tudo, got %+v", stats)
	}
}

func TestCalculate_TotalAndActiveDays(t *testing.T) {
	days := []ContributionDay{
		mkDay(4, 0),
		mkDay(3, 2),
		mkDay(2, 0),
		mkDay(1, 7),
		mkDay(0, 1),
	}
	stats := Calculate(days)
	if stats.TotalDays != 5 {
		t.Errorf("TotalDays = %d, quer 5", stats.TotalDays)
	}
	if stats.ActiveDays != 3 {
		t.Errorf("ActiveDays = %d, quer 3", stats.ActiveDays)
	}
}

func TestCalculate_LongestStreak(t *testing.T) {
	// run de 3 dias, gap, run de 2 (terminando hoje)
	days := []ContributionDay{
		mkDay(5, 1),
		mkDay(4, 1),
		mkDay(3, 1),
		mkDay(2, 0), // gap quebra o run
		mkDay(1, 1),
		mkDay(0, 1),
	}
	stats := Calculate(days)
	if stats.LongestStreak != 3 {
		t.Errorf("LongestStreak = %d, quer 3", stats.LongestStreak)
	}
}

func TestCalculate_LongestPicksBiggestRun(t *testing.T) {
	// run de 2, gap, run de 4 -> longest deve ser 4
	days := []ContributionDay{
		mkDay(6, 1), mkDay(5, 1),
		mkDay(4, 0),
		mkDay(3, 1), mkDay(2, 1), mkDay(1, 1), mkDay(0, 1),
	}
	stats := Calculate(days)
	if stats.LongestStreak != 4 {
		t.Errorf("LongestStreak = %d, quer 4", stats.LongestStreak)
	}
}

func TestCalculate_CurrentStreak(t *testing.T) {
	days := []ContributionDay{
		mkDay(2, 5),
		mkDay(1, 3),
		mkDay(0, 2), // hoje com commit
	}
	if got := Calculate(days).CurrentStreak; got != 3 {
		t.Errorf("CurrentStreak = %d, quer 3", got)
	}
}

func TestCalculate_CurrentStreakStopsAtGap(t *testing.T) {
	days := []ContributionDay{
		mkDay(2, 5),
		mkDay(1, 0), // ontem sem commit quebra
		mkDay(0, 2), // hoje com commit
	}
	if got := Calculate(days).CurrentStreak; got != 1 {
		t.Errorf("CurrentStreak = %d, quer 1 (só hoje)", got)
	}
}

func TestCalculate_TodayWithoutCommitDoesNotBreak(t *testing.T) {
	// Regra do GitHub: hoje ainda sem commit não derruba o streak.
	days := []ContributionDay{
		mkDay(2, 5),
		mkDay(1, 3),
		mkDay(0, 0), // hoje ainda sem commit
	}
	if got := Calculate(days).CurrentStreak; got != 2 {
		t.Errorf("CurrentStreak = %d, quer 2 (hoje 0 não quebra)", got)
	}
}

func TestIsToday(t *testing.T) {
	if !isToday(time.Now()) {
		t.Error("isToday(agora) deveria ser true")
	}
	if isToday(time.Now().AddDate(0, 0, -1)) {
		t.Error("isToday(ontem) deveria ser false")
	}
}
