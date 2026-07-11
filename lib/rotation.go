package lib

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// rotationWorkerCount bounds concurrent CreateChore calls when generating a rotation.
const rotationWorkerCount = 5

// CreateChoreRotation generates rotating chore assignments across family members over N weeks.
// Creates are parallelized with a bounded worker pool (rotationWorkerCount).
func (c *Client) CreateChoreRotation(frameID string, data RotationData) (*RotationResult, error) {
	if len(data.Chores) == 0 {
		return nil, errors.New("at least one chore is required")
	}
	if len(data.AssigneeIDs) == 0 {
		return nil, errors.New("at least one assignee is required")
	}
	if data.Weeks <= 0 {
		return nil, errors.New("weeks must be greater than zero")
	}

	startDate, err := time.Parse(DateFormat, data.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format (expected YYYY-MM-DD): %w", err)
	}

	type job struct {
		title    string
		dueDate  string
		assignee string
		week     int
	}

	numAssignees := len(data.AssigneeIDs)
	jobs := make([]job, 0, data.Weeks*len(data.Chores))
	for week := 0; week < data.Weeks; week++ {
		dueDate := startDate.AddDate(0, 0, week*7).Format(DateFormat)
		for choreIdx, choreTitle := range data.Chores {
			assignee := data.AssigneeIDs[(choreIdx+week)%numAssignees]
			jobs = append(jobs, job{
				title:    choreTitle,
				dueDate:  dueDate,
				assignee: assignee,
				week:     week + 1,
			})
		}
	}

	type outcome struct {
		chore *Chore
		err   error
	}

	out := make(chan outcome, len(jobs))
	sem := make(chan struct{}, rotationWorkerCount)
	var wg sync.WaitGroup

	for _, j := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func(j job) {
			defer wg.Done()
			defer func() { <-sem }()
			chore, err := c.CreateChore(frameID, ChoreData{
				Title:      j.title,
				DueDate:    j.dueDate,
				Points:     data.Points,
				AssigneeID: j.assignee,
			})
			if err != nil {
				out <- outcome{err: fmt.Errorf("failed creating chore %q for week %d: %w", j.title, j.week, err)}
				return
			}
			out <- outcome{chore: chore}
		}(j)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	var created []Chore
	var firstErr error
	for o := range out {
		if o.err != nil {
			if firstErr == nil {
				firstErr = o.err
			}
			continue
		}
		if o.chore != nil {
			created = append(created, *o.chore)
		}
	}

	if firstErr != nil {
		return &RotationResult{Chores: created}, firstErr
	}
	return &RotationResult{Chores: created}, nil
}
