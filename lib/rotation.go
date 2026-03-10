package lib

import (
	"errors"
	"fmt"
	"time"
)

// CreateChoreRotation generates rotating chore assignments across family members over N weeks.
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

	var created []Chore
	numAssignees := len(data.AssigneeIDs)

	for week := 0; week < data.Weeks; week++ {
		dueDate := startDate.AddDate(0, 0, week*7).Format(DateFormat)

		for choreIdx, choreTitle := range data.Chores {
			assignee := data.AssigneeIDs[(choreIdx+week)%numAssignees]

			chore, err := c.CreateChore(frameID, ChoreData{
				Title:      choreTitle,
				DueDate:    dueDate,
				Points:     data.Points,
				AssigneeID: assignee,
			})
			if err != nil {
				return &RotationResult{Chores: created}, fmt.Errorf("failed creating chore %q for week %d: %w", choreTitle, week+1, err)
			}

			created = append(created, *chore)
		}
	}

	return &RotationResult{Chores: created}, nil
}
