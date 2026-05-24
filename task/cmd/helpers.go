package cmd

import (
	"sort"
	"time"

	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/scoring"
)

type scoredTask struct {
	task  db.Task
	score scoring.ScoreResult
}

func scoreAndSort(tasks []db.Task, database *db.DB, currentBranch string, weights scoring.WeightMap) ([]scoredTask, error) {
	now := time.Now()
	recentShown, err := database.GetRecentShown()
	if err != nil {
		return nil, err
	}

	results := make([]scoredTask, len(tasks))
	for i, task := range tasks {
		cnt, _ := database.BlockedByCount(task.ID)
		in := scoring.Input{
			TaskID:         task.ID,
			DueDate:        task.DueDate,
			BlockedByCount: cnt,
			Branch:         task.Branch,
			CurrentBranch:  currentBranch,
			CreatedAt:      task.CreatedAt,
			Now:            now,
			RecentShown:    recentShown,
			Weights:        weights,
		}
		results[i] = scoredTask{task, scoring.Calculate(in)}
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].task.Pinned != results[j].task.Pinned {
			return results[i].task.Pinned
		}
		return results[i].score.Total > results[j].score.Total
	})

	return results, nil
}
