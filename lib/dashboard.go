package lib

import (
	"sync"
	"time"
)

// DateFormat is the standard date layout used across the Skylight API (YYYY-MM-DD).
const DateFormat = "2006-01-02"

// GetDashboard aggregates today's events, chores, points, meals, and lists.
func (c *Client) GetDashboard(frameID string) (*Dashboard, error) {
	today := time.Now().Format(DateFormat)

	var (
		events   []CalendarEvent
		chores   []Chore
		points   []RewardPointEntry
		sittings []MealSitting
		lists    []List
		errs     [5]error
		wg       sync.WaitGroup
	)

	wg.Add(5)
	go func() {
		defer wg.Done()
		events, errs[0] = c.ListCalendarEvents(frameID, today, today)
	}()
	go func() {
		defer wg.Done()
		chores, errs[1] = c.ListChores(frameID, ChoreListOptions{
			Date:        today,
			Status:      choreStatusPending,
			IncludeLate: true,
		})
	}()
	go func() {
		defer wg.Done()
		points, errs[2] = c.GetRewardPoints(frameID)
	}()
	go func() {
		defer wg.Done()
		sittings, errs[3] = c.ListMealSittings(frameID, MealSittingListOptions{
			DateMin: today,
			DateMax: today,
		})
	}()
	go func() {
		defer wg.Done()
		lists, errs[4] = c.ListLists(frameID)
	}()
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return &Dashboard{
		Date:         today,
		Events:       events,
		Chores:       chores,
		Points:       points,
		MealSittings: sittings,
		Lists:        lists,
	}, nil
}
