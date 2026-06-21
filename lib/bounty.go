package lib

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// CreateBounty creates a chore and a paired reward as a single bounty.
// If reward creation fails, a best-effort cleanup deletes the chore.
func (c *Client) CreateBounty(frameID string, data BountyData) (*Bounty, error) {
	chore, err := c.CreateChore(frameID, ChoreData{
		Title:      data.Title,
		Points:     data.Points,
		DueDate:    data.DueDate,
		AssigneeID: data.AssigneeID,
		Recurring:  data.Recurring,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bounty chore: %w", err)
	}

	categoryIDs := data.CategoryIDs
	if len(categoryIDs) == 0 && data.AssigneeID != "" {
		id, err := strconv.Atoi(data.AssigneeID)
		if err != nil {
			choreDelErr := c.DeleteChore(frameID, chore.ID)
			if choreDelErr != nil {
				fmt.Fprintf(os.Stderr, "WARN: bounty cleanup delete chore failed: %v\n", choreDelErr)
			}
			return nil, fmt.Errorf("assignee_id %q is not a valid numeric category ID: %w", data.AssigneeID, err)
		}
		categoryIDs = []int{id}
	}
	reward, err := c.CreateReward(frameID, RewardData{
		Title:       data.RewardTitle,
		Points:      data.Points,
		EmojiIcon:   data.EmojiIcon,
		CategoryIDs: categoryIDs,
	})
	if err != nil {
		// Best-effort cleanup
		choreDelErr := c.DeleteChore(frameID, chore.ID)
		if choreDelErr != nil {
			fmt.Fprintf(os.Stderr, "WARN: bounty cleanup delete chore failed: %v\n", choreDelErr)
		}
		return nil, fmt.Errorf("failed to create bounty reward: %w", err)
	}

	return &Bounty{
		Chore:  *chore,
		Reward: *reward,
	}, nil
}

// DeleteBounty deletes the chore and reward that make up a bounty.
// Both deletions are attempted; the first error encountered is returned.
func (c *Client) DeleteBounty(frameID, choreID, rewardID string) error {
	choreErr := c.DeleteChore(frameID, choreID)
	rewardErr := c.DeleteReward(frameID, rewardID)
	if choreErr != nil {
		return fmt.Errorf("failed to delete bounty chore: %w", choreErr)
	}
	if rewardErr != nil {
		return fmt.Errorf("failed to delete bounty reward: %w", rewardErr)
	}
	return nil
}

// UpdateBounty updates the chore and reward that make up a bounty.
// Only fields set in data are applied; zero-value fields are ignored by the underlying update calls.
func (c *Client) UpdateBounty(frameID, choreID, rewardID string, data BountyData) (*Bounty, error) {
	chore, err := c.UpdateChore(frameID, choreID, ChoreData{
		Title:   data.Title,
		Points:  data.Points,
		DueDate: data.DueDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update bounty chore: %w", err)
	}

	reward, err := c.UpdateReward(frameID, rewardID, RewardData{
		Title:     data.RewardTitle,
		Points:    data.Points,
		EmojiIcon: data.EmojiIcon,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update bounty reward: %w", err)
	}

	return &Bounty{Chore: *chore, Reward: *reward}, nil
}

// ListBounties lists pending chores with points and unredeemed rewards,
// matching them by point value as a heuristic.
func (c *Client) ListBounties(frameID string) ([]Bounty, error) {
	today := time.Now()

	var (
		chores    []Chore
		rewards   []Reward
		choreErr  error
		rewardErr error
		wg        sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		chores, choreErr = c.ListChores(frameID, ChoreListOptions{
			Status: choreStatusPending,
			After:  today.AddDate(0, 0, -1).Format(DateFormat),
			Before: today.AddDate(0, 1, 0).Format(DateFormat),
		})
	}()
	go func() {
		defer wg.Done()
		rewards, rewardErr = c.ListRewards(frameID)
	}()
	wg.Wait()

	if choreErr != nil {
		return nil, fmt.Errorf("failed to list bounty chores: %w", choreErr)
	}
	if rewardErr != nil {
		return nil, fmt.Errorf("failed to list bounty rewards: %w", rewardErr)
	}

	// Index unredeemed rewards by point value
	rewardsByPoints := map[int][]Reward{}
	for _, r := range rewards {
		if !r.Redeemed {
			rewardsByPoints[r.Points] = append(rewardsByPoints[r.Points], r)
		}
	}

	var bounties []Bounty
	for _, ch := range chores {
		if ch.Points <= 0 {
			continue
		}
		if matches, ok := rewardsByPoints[ch.Points]; ok && len(matches) > 0 {
			bounties = append(bounties, Bounty{
				Chore:  ch,
				Reward: matches[0],
			})
			rewardsByPoints[ch.Points] = matches[1:]
		}
	}

	return bounties, nil
}
