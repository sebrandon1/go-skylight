package lib

import "fmt"

// ListRewards retrieves rewards for a frame.
func (c *Client) ListRewards(frameID string) ([]Reward, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/rewards", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list rewards request: %w", err)
	}

	var apiResp rewardAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list rewards: %w", err)
	}

	rewards := make([]Reward, len(apiResp.Data))
	for i := range apiResp.Data {
		rewards[i] = apiResp.Data[i].toReward()
	}

	return rewards, nil
}

// CreateReward creates a new reward on a frame.
func (c *Client) CreateReward(frameID string, reward RewardData) (*Reward, error) {
	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/rewards", SkylightURL, frameID), reward)
	if err != nil {
		return nil, fmt.Errorf("failed to create reward request: %w", err)
	}

	var apiResp rewardAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create reward: %w", err)
	}

	result := apiResp.Data.toReward()
	return &result, nil
}

// UpdateReward updates an existing reward.
func (c *Client) UpdateReward(frameID, rewardID string, reward RewardData) (*Reward, error) {
	req, err := newRequestWithBody("PATCH", fmt.Sprintf("%s/frames/%s/rewards/%s", SkylightURL, frameID, rewardID), reward)
	if err != nil {
		return nil, fmt.Errorf("failed to create update reward request: %w", err)
	}

	var apiResp rewardAPISingleResponse
	if err := c.patch(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update reward: %w", err)
	}

	result := apiResp.Data.toReward()
	return &result, nil
}

// DeleteReward deletes a reward.
func (c *Client) DeleteReward(frameID, rewardID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/rewards/%s", SkylightURL, frameID, rewardID), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete reward request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete reward: %w", err)
	}

	return nil
}

// RedeemReward redeems a reward.
func (c *Client) RedeemReward(frameID, rewardID string) error {
	req, err := newRequest("POST", fmt.Sprintf("%s/frames/%s/rewards/%s/redeem", SkylightURL, frameID, rewardID), nil)
	if err != nil {
		return fmt.Errorf("failed to create redeem reward request: %w", err)
	}

	if err := c.post(req, nil); err != nil {
		return fmt.Errorf("failed to redeem reward: %w", err)
	}

	return nil
}

// UnredeemReward unredeems a reward.
func (c *Client) UnredeemReward(frameID, rewardID string) error {
	req, err := newRequest("POST", fmt.Sprintf("%s/frames/%s/rewards/%s/unredeem", SkylightURL, frameID, rewardID), nil)
	if err != nil {
		return fmt.Errorf("failed to create unredeem reward request: %w", err)
	}

	if err := c.post(req, nil); err != nil {
		return fmt.Errorf("failed to unredeem reward: %w", err)
	}

	return nil
}

// GetRewardPoints retrieves reward points for a frame.
func (c *Client) GetRewardPoints(frameID string) ([]RewardPointEntry, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/reward_points", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get reward points request: %w", err)
	}

	var points []RewardPointEntry
	if err := c.get(req, &points); err != nil {
		return nil, fmt.Errorf("failed to get reward points: %w", err)
	}

	return points, nil
}
