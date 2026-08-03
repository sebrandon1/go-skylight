package lib

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// CreateBounty creates a chore and a paired reward as a single bounty.
// If reward creation fails, a best-effort cleanup deletes the chore.
func (c *Client) CreateBounty(ctx context.Context, frameID string, data BountyData) (*Bounty, error) {
	chore, err := c.CreateChore(ctx, frameID, ChoreData{
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
			if delErr := c.DeleteChore(ctx, frameID, chore.ID); delErr != nil && c.logger != nil {
				c.logger.Warn("bounty cleanup: failed to delete chore after assignee parse error", "chore_id", chore.ID, "error", delErr)
			}
			return nil, fmt.Errorf("assignee_id %q is not a valid numeric category ID: %w", data.AssigneeID, err)
		}
		categoryIDs = []int{id}
	}
	reward, err := c.CreateReward(ctx, frameID, RewardData{
		Title:       data.RewardTitle,
		Points:      data.Points,
		EmojiIcon:   data.EmojiIcon,
		CategoryIDs: categoryIDs,
	})
	if err != nil {
		if delErr := c.DeleteChore(ctx, frameID, chore.ID); delErr != nil && c.logger != nil {
			c.logger.Warn("bounty cleanup: failed to delete chore after reward creation error", "chore_id", chore.ID, "error", delErr)
		}
		return nil, fmt.Errorf("failed to create bounty reward: %w", err)
	}

	return &Bounty{
		Chore:  *chore,
		Reward: *reward,
	}, nil
}

// DeleteBounty deletes the chore and reward that make up a bounty.
// Both deletions are always attempted; any errors are joined and returned together.
func (c *Client) DeleteBounty(ctx context.Context, frameID, choreID, rewardID string) error {
	choreErr := c.DeleteChore(ctx, frameID, choreID)
	rewardErr := c.DeleteReward(ctx, frameID, rewardID)
	if choreErr != nil {
		choreErr = fmt.Errorf("failed to delete bounty chore: %w", choreErr)
	}
	if rewardErr != nil {
		rewardErr = fmt.Errorf("failed to delete bounty reward: %w", rewardErr)
	}
	return errors.Join(choreErr, rewardErr)
}

// UpdateBounty updates the chore and reward that make up a bounty.
// Only fields set in data are applied; zero-value fields are ignored by the underlying update calls.
func (c *Client) UpdateBounty(ctx context.Context, frameID, choreID, rewardID string, data BountyData) (*Bounty, error) {
	chore, err := c.UpdateChore(ctx, frameID, choreID, ChoreData{
		Title:   data.Title,
		Points:  data.Points,
		DueDate: data.DueDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update bounty chore: %w", err)
	}

	reward, err := c.UpdateReward(ctx, frameID, rewardID, RewardData{
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
func (c *Client) ListBounties(ctx context.Context, frameID string) ([]Bounty, error) {
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
		chores, choreErr = c.ListChores(ctx, frameID, ChoreListOptions{
			Status: choreStatusPending,
			After:  today.AddDate(0, 0, -1).Format(DateFormat),
			Before: today.AddDate(0, 1, 0).Format(DateFormat),
		})
	}()
	go func() {
		defer wg.Done()
		rewards, rewardErr = c.ListRewards(ctx, frameID)
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
