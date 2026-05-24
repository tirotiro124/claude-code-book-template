package scoring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	now      = time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	balanced = PresetWeights[PresetBalanced]
)

func day(d int) *time.Time {
	t := now.AddDate(0, 0, d)
	return &t
}

func TestCalcUrgency(t *testing.T) {
	tests := []struct {
		name     string
		dueDate  *time.Time
		expected int
	}{
		{"締め切りなし", nil, 0},
		{"8日後", day(8), 0},
		{"7日後", day(7), 15},
		{"4日後", day(4), 15},
		{"3日後", day(3), 30},
		{"1日後", day(1), 30},
		{"当日", day(0), 40},
		{"1日超過", day(-1), 42},
		{"3日超過", day(-3), 46},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, calcUrgency(tt.dueDate, now))
		})
	}
}

func TestCalcImportance(t *testing.T) {
	tests := []struct {
		count    int
		expected int
	}{
		{0, 0},
		{1, 10},
		{2, 20},
		{3, 30},
		{5, 30},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, calcImportance(tt.count))
	}
}

func TestCalcContext(t *testing.T) {
	tests := []struct {
		name          string
		taskBranch    string
		currentBranch string
		expected      int
	}{
		{"ブランチ一致", "feat/login", "feat/login", 15},
		{"ブランチ不一致", "feat/login", "main", 0},
		{"タスクブランチが空", "", "main", 0},
		{"両方空", "", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, calcContext(tt.taskBranch, tt.currentBranch))
		})
	}
}

func TestCalcAging(t *testing.T) {
	tests := []struct {
		name     string
		daysAgo  int
		expected int
	}{
		{"2日前に作成", 2, 0},
		{"3日前に作成", 3, 5},
		{"6日前に作成", 6, 5},
		{"7日前に作成", 7, 10},
		{"13日前に作成", 13, 10},
		{"14日前に作成", 14, 15},
		{"30日前に作成", 30, 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdAt := now.AddDate(0, 0, -tt.daysAgo)
			assert.Equal(t, tt.expected, calcAging(createdAt, now))
		})
	}
}

func TestCalcFatigue(t *testing.T) {
	tests := []struct {
		name        string
		taskID      int64
		recentShown []int64
		expected    int
	}{
		{"直前に表示済み", 1, []int64{1, 2}, 10},
		{"2回前に表示済み", 2, []int64{1, 2}, 5},
		{"表示履歴なし", 3, []int64{1, 2}, 0},
		{"履歴が空", 1, []int64{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, calcFatigue(tt.taskID, tt.recentShown))
		})
	}
}

func TestCalculate_Balanced(t *testing.T) {
	input := Input{
		DueDate:        day(2),
		BlockedByCount: 3,
		Branch:         "feat/login",
		CurrentBranch:  "feat/login",
		CreatedAt:      now.AddDate(0, 0, -5),
		Now:            now,
		RecentShown:    []int64{99, 88},
		Weights:        balanced,
	}
	result := Calculate(input)
	assert.Equal(t, 30, result.Urgency)    // 2日後
	assert.Equal(t, 30, result.Importance) // 3件以上
	assert.Equal(t, 15, result.Context)    // ブランチ一致
	assert.Equal(t, 5, result.Aging)       // 5日前作成（3〜6日 → 5点）
	assert.Equal(t, 0, result.Fatigue)     // 履歴になし
	assert.Equal(t, 80, result.Total)
}

func TestCalculate_WithFatigue(t *testing.T) {
	input := Input{
		CreatedAt:   now.AddDate(0, 0, -1),
		Now:         now,
		RecentShown: []int64{42, 0},
		Weights:     balanced,
		TaskID:      42,
	}
	result := Calculate(input)
	assert.Equal(t, 10, result.Fatigue)
	assert.Equal(t, -10, result.Total)
}

func TestCalculate_DeadlinePreset(t *testing.T) {
	weights := PresetWeights[PresetDeadline]
	input := Input{
		DueDate:     day(2),
		CreatedAt:   now,
		Now:         now,
		RecentShown: []int64{},
		Weights:     weights,
	}
	result := Calculate(input)
	// urgency(30) * 2.0 = 60
	assert.Equal(t, 60, result.Total)
}

func TestCalculate_NilWeightsFallsBackToBalanced(t *testing.T) {
	input := Input{
		DueDate:     day(2),
		CreatedAt:   now,
		Now:         now,
		RecentShown: []int64{},
		Weights:     nil, // nil → balanced にフォールバック
	}
	result := Calculate(input)
	assert.Equal(t, 30, result.Urgency)
	assert.Equal(t, 30, result.Total)
}

func TestCalculate_FlowPreset(t *testing.T) {
	weights := PresetWeights[PresetFlow]
	input := Input{
		Branch:        "feat/login",
		CurrentBranch: "feat/login",
		CreatedAt:     now,
		Now:           now,
		RecentShown:   []int64{},
		Weights:       weights,
	}
	result := Calculate(input)
	// context(15) * 2.0 = 30
	assert.Equal(t, 30, result.Total)
}

func BenchmarkCalculate(b *testing.B) {
	due := now.AddDate(0, 0, 2)
	input := Input{
		TaskID:         1,
		DueDate:        &due,
		BlockedByCount: 2,
		Branch:         "feat/login",
		CurrentBranch:  "feat/login",
		CreatedAt:      now.AddDate(0, 0, -10),
		Now:            now,
		RecentShown:    []int64{2, 3},
		Weights:        PresetWeights[PresetBalanced],
	}
	for b.Loop() {
		Calculate(input)
	}
}
