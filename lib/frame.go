package lib

import "fmt"

// GetFrame retrieves frame information.
func (c *Client) GetFrame(frameID string) (*Frame, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get frame request: %w", err)
	}

	var frame Frame
	if err := c.get(req, &frame); err != nil {
		return nil, fmt.Errorf("failed to get frame: %w", err)
	}

	return &frame, nil
}

// ListDevices retrieves devices for a frame.
func (c *Client) ListDevices(frameID string) ([]Device, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/devices", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list devices request: %w", err)
	}

	var devices []Device
	if err := c.get(req, &devices); err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	return devices, nil
}

// GetAvatars retrieves available avatars.
func (c *Client) GetAvatars() ([]Avatar, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/avatars", SkylightURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get avatars request: %w", err)
	}

	var avatars []Avatar
	if err := c.get(req, &avatars); err != nil {
		return nil, fmt.Errorf("failed to get avatars: %w", err)
	}

	return avatars, nil
}

// GetColors retrieves available colors.
func (c *Client) GetColors() ([]Color, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/colors", SkylightURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get colors request: %w", err)
	}

	var colors []Color
	if err := c.get(req, &colors); err != nil {
		return nil, fmt.Errorf("failed to get colors: %w", err)
	}

	return colors, nil
}
