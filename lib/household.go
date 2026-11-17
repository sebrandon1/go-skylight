package lib

import (
	"context"
	"fmt"
)

// HouseholdConfig holds the user-adjustable add-on settings of a frame's
// household. Today these are the Disney Mode add-on's switches. The add-ons a
// frame is entitled to are reported separately, in Frame.FeatureBundle, and
// cannot be turned on or off through the API.
type HouseholdConfig struct {
	DisneyProfilePictures bool `json:"disney_profile_pictures"`
	DisneyScreensaver     bool `json:"disney_screensaver"`
}

// HouseholdConfigUpdate is a partial update; nil fields are left unchanged.
type HouseholdConfigUpdate struct {
	DisneyProfilePictures *bool `json:"disney_profile_pictures,omitempty"`
	DisneyScreensaver     *bool `json:"disney_screensaver,omitempty"`
}

type householdConfigResponse struct {
	Data struct {
		Attributes HouseholdConfig `json:"attributes"`
	} `json:"data"`
}

func (c *Client) householdConfigURL(frameID string) string {
	return fmt.Sprintf("%s/frames/%s/household_config", c.effectiveURL(), pathSeg(frameID))
}

// GetHouseholdConfig retrieves the frame's household add-on settings.
func (c *Client) GetHouseholdConfig(ctx context.Context, frameID string) (*HouseholdConfig, error) {
	req, err := newRequest(ctx, "GET", c.householdConfigURL(frameID))
	if err != nil {
		return nil, fmt.Errorf("failed to create get household config request: %w", err)
	}
	var resp householdConfigResponse
	if err := c.get(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get household config: %w", err)
	}
	return &resp.Data.Attributes, nil
}

// UpdateHouseholdConfig changes the given household add-on settings and
// returns the settings as stored afterwards.
func (c *Client) UpdateHouseholdConfig(ctx context.Context, frameID string, update HouseholdConfigUpdate) (*HouseholdConfig, error) {
	if update.DisneyProfilePictures == nil && update.DisneyScreensaver == nil {
		return nil, fmt.Errorf("no household config settings to update")
	}
	req, err := newRequestWithBody(ctx, "PATCH", c.householdConfigURL(frameID), update)
	if err != nil {
		return nil, fmt.Errorf("failed to create update household config request: %w", err)
	}
	var resp householdConfigResponse
	if err := c.patch(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to update household config: %w", err)
	}
	return &resp.Data.Attributes, nil
}
