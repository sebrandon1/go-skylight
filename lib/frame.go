package lib

import (
	"fmt"
)

// ListFrames retrieves all frames accessible to the authenticated user.
func (c *Client) ListFrames() ([]Frame, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames", c.effectiveURL()))
	if err != nil {
		return nil, fmt.Errorf("failed to create list frames request: %w", err)
	}

	var apiResp framesAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list frames: %w", err)
	}

	frames := make([]Frame, len(apiResp.Data))
	for i := range apiResp.Data {
		frames[i] = apiResp.Data[i].toFrame()
	}
	return frames, nil
}

// GetFrame retrieves frame information.
func (c *Client) GetFrame(frameID string) (*Frame, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s", c.effectiveURL(), frameID))
	if err != nil {
		return nil, fmt.Errorf("failed to create get frame request: %w", err)
	}

	var apiResp frameAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to get frame: %w", err)
	}

	result := apiResp.Data.toFrame()
	return &result, nil
}

// ListDevices retrieves devices for a frame.
func (c *Client) ListDevices(frameID string) ([]Device, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/devices", c.effectiveURL(), frameID))
	if err != nil {
		return nil, fmt.Errorf("failed to create list devices request: %w", err)
	}

	var apiResp deviceAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	devices := make([]Device, len(apiResp.Data))
	for i := range apiResp.Data {
		devices[i] = apiResp.Data[i].toDevice()
	}
	return devices, nil
}

// GetAvatars retrieves available avatars.
func (c *Client) GetAvatars() ([]Avatar, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/avatars", c.effectiveURL()))
	if err != nil {
		return nil, fmt.Errorf("failed to create get avatars request: %w", err)
	}

	var apiResp avatarAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to get avatars: %w", err)
	}

	avatars := make([]Avatar, len(apiResp.Data))
	for i := range apiResp.Data {
		avatars[i] = apiResp.Data[i].toAvatar()
	}
	return avatars, nil
}

// SetCurrentAlbum sets the active slideshow album for a frame by album ID.
// Use -1 to revert to the default (all photos) album.
func (c *Client) SetCurrentAlbum(frameID string, albumID int) error {
	body := struct {
		Frame struct {
			CurrentAlbumID int `json:"current_album_id"`
		} `json:"frame"`
	}{}
	body.Frame.CurrentAlbumID = albumID

	req, err := newRequestWithBody("PATCH", fmt.Sprintf("%s/frames/%s", c.effectiveURL(), frameID), body)
	if err != nil {
		return fmt.Errorf("failed to create set album request: %w", err)
	}

	if err := c.patch(req, nil); err != nil {
		return fmt.Errorf("failed to set current album: %w", err)
	}
	return nil
}

// GetColors retrieves available colors.
func (c *Client) GetColors() ([]Color, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/colors", c.effectiveURL()))
	if err != nil {
		return nil, fmt.Errorf("failed to create get colors request: %w", err)
	}

	var apiResp colorAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to get colors: %w", err)
	}

	return apiResp.Data, nil
}
