package lib

import (
	"time"
)

// DateFormat is the standard date layout used across the Skylight API (YYYY-MM-DD).
const DateFormat = "2006-01-02"

// GetDashboard aggregates today's events, chores, points, meals, and lists.
func (c *Client) GetDashboard(frameID string) (*Dashboard, error) {
	today := time.Now().Format(DateFormat)

	events, err := c.ListCalendarEvents(frameID, today, today)
	if err != nil {
		return nil, err
	}

	chores, err := c.ListChores(frameID, ChoreListOptions{
		Date:        today,
		Status:      "pending",
		IncludeLate: true,
	})
	if err != nil {
		return nil, err
	}

	points, err := c.GetRewardPoints(frameID)
	if err != nil {
		return nil, err
	}

	allSittings, err := c.ListMealSittings(frameID)
	if err != nil {
		return nil, err
	}

	var todaySittings []MealSitting
	for _, s := range allSittings {
		if s.Date == today {
			todaySittings = append(todaySittings, s)
		}
	}

	lists, err := c.ListLists(frameID)
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		Date:         today,
		Events:       events,
		Chores:       chores,
		Points:       points,
		MealSittings: todaySittings,
		Lists:        lists,
	}, nil
}
