package lib

import "fmt"

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

	reward, err := c.CreateReward(frameID, RewardData{
		Title:     data.RewardTitle,
		Points:    data.Points,
		EmojiIcon: data.EmojiIcon,
	})
	if err != nil {
		// Best-effort cleanup
		_ = c.DeleteChore(frameID, chore.ID)
		return nil, fmt.Errorf("failed to create bounty reward: %w", err)
	}

	return &Bounty{
		Chore:  *chore,
		Reward: *reward,
	}, nil
}

// ListBounties lists pending chores with points and unredeemed rewards,
// matching them by point value as a heuristic.
func (c *Client) ListBounties(frameID string) ([]Bounty, error) {
	chores, err := c.ListChores(frameID, ChoreListOptions{Status: "pending"})
	if err != nil {
		return nil, fmt.Errorf("failed to list bounty chores: %w", err)
	}

	rewards, err := c.ListRewards(frameID)
	if err != nil {
		return nil, fmt.Errorf("failed to list bounty rewards: %w", err)
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
