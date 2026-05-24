package scoring

import (
	"math"
	"time"
)

// Preset はスコア重み係数セットの識別子。
type Preset string

const (
	PresetDeadline Preset = "deadline"
	PresetBalanced Preset = "balanced"
	PresetFlow     Preset = "flow"
)

// WeightMap は各シグナルの重み係数を保持する。
type WeightMap map[string]float64

// PresetWeights は各プリセットの重み係数定義。
var PresetWeights = map[Preset]WeightMap{
	PresetDeadline: {"urgency": 2.0, "importance": 1.0, "context": 0.5, "aging": 1.0, "fatigue": 1.0},
	PresetBalanced: {"urgency": 1.0, "importance": 1.0, "context": 1.0, "aging": 1.0, "fatigue": 1.0},
	PresetFlow:     {"urgency": 0.5, "importance": 1.0, "context": 2.0, "aging": 1.0, "fatigue": 1.0},
}

// Input はスコア計算に必要なすべての入力値。
type Input struct {
	TaskID         int64
	DueDate        *time.Time
	BlockedByCount int
	Branch         string
	CurrentBranch  string
	CreatedAt      time.Time
	Now            time.Time
	RecentShown    []int64
	Weights        WeightMap
}

// ScoreResult はスコア計算結果と各シグナルの内訳。
type ScoreResult struct {
	Total      int
	Urgency    int
	Importance int
	Context    int
	Aging      int
	Fatigue    int
}

// Calculate はタスクのスコアを計算する。副作用を持たない純粋関数。
func Calculate(in Input) ScoreResult {
	w := in.Weights
	if w == nil {
		w = PresetWeights[PresetBalanced]
	}

	urgency := calcUrgency(in.DueDate, in.Now)
	importance := calcImportance(in.BlockedByCount)
	context := calcContext(in.Branch, in.CurrentBranch)
	aging := calcAging(in.CreatedAt, in.Now)
	fatigue := calcFatigue(in.TaskID, in.RecentShown)

	total := int(math.Round(float64(urgency)*w["urgency"])) +
		int(math.Round(float64(importance)*w["importance"])) +
		int(math.Round(float64(context)*w["context"])) +
		int(math.Round(float64(aging)*w["aging"])) -
		int(math.Round(float64(fatigue)*w["fatigue"]))

	return ScoreResult{
		Total:      total,
		Urgency:    urgency,
		Importance: importance,
		Context:    context,
		Aging:      aging,
		Fatigue:    fatigue,
	}
}

func calcUrgency(dueDate *time.Time, now time.Time) int {
	if dueDate == nil {
		return 0
	}
	days := int(dueDate.Truncate(24*time.Hour).Sub(now.Truncate(24*time.Hour)).Hours() / 24)
	switch {
	case days >= 8:
		return 0
	case days >= 4:
		return 15
	case days >= 1:
		return 30
	case days == 0:
		return 40
	default: // 超過
		return 40 + (-days)*2
	}
}

func calcImportance(blockedByCount int) int {
	switch {
	case blockedByCount <= 0:
		return 0
	case blockedByCount == 1:
		return 10
	case blockedByCount == 2:
		return 20
	default:
		return 30
	}
}

func calcContext(taskBranch, currentBranch string) int {
	if taskBranch != "" && taskBranch == currentBranch {
		return 15
	}
	return 0
}

func calcAging(createdAt, now time.Time) int {
	days := int(now.Truncate(24*time.Hour).Sub(createdAt.Truncate(24*time.Hour)).Hours() / 24)
	switch {
	case days < 3:
		return 0
	case days < 7:
		return 5
	case days < 14:
		return 10
	default:
		return 15
	}
}

func calcFatigue(taskID int64, recentShown []int64) int {
	if len(recentShown) > 0 && recentShown[0] == taskID {
		return 10
	}
	if len(recentShown) > 1 && recentShown[1] == taskID {
		return 5
	}
	return 0
}
