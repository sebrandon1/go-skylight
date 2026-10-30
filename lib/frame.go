package lib

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// ListFrames retrieves all frames accessible to the authenticated user.
func (c *Client) ListFrames(ctx context.Context) ([]Frame, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames", c.effectiveURL()))
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
func (c *Client) GetFrame(ctx context.Context, frameID string) (*Frame, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s", c.effectiveURL(), pathSeg(frameID)))
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
func (c *Client) ListDevices(ctx context.Context, frameID string) ([]Device, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/devices", c.effectiveURL(), pathSeg(frameID)))
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

// tabAttrSuffix is appended to a tab name to form its device_config attribute
// (e.g. "meals" -> "meals_tab").
const tabAttrSuffix = "_tab"

// hideableTabs lists the navigation tabs a device can show or hide, in the
// order the Skylight app lists them.
var hideableTabs = []string{"home", "calendar", "lists", "tasks", "rewards", "meals", "recipes", "photos", "sleep"}

// HideableTabs returns the tab names accepted by UpdateDeviceTabs, in the order
// the Skylight app lists them.
func HideableTabs() []string {
	return slices.Clone(hideableTabs)
}

// GetDeviceConfig retrieves the navigation tab visibility of a device.
func (c *Client) GetDeviceConfig(ctx context.Context, frameID, deviceID string) (*DeviceConfig, error) {
	req, err := newRequest(ctx, "GET", c.deviceConfigURL(frameID, deviceID))
	if err != nil {
		return nil, fmt.Errorf("failed to create get device config request: %w", err)
	}

	var apiResp deviceConfigAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to get device config: %w", err)
	}

	cfg := apiResp.toDeviceConfig(deviceID)
	return &cfg, nil
}

// UpdateDeviceTabs shows (true) or hides (false) navigation tabs on a device.
// Keys are tab names from HideableTabs; tabs not in the map are left unchanged.
// Hiding a tab does not delete any of its data.
func (c *Client) UpdateDeviceTabs(ctx context.Context, frameID, deviceID string, tabs map[string]bool) (*DeviceConfig, error) {
	if len(tabs) == 0 {
		return nil, fmt.Errorf("no tabs to update")
	}
	body := make(map[string]bool, len(tabs))
	for name, shown := range tabs {
		if !slices.Contains(hideableTabs, name) {
			return nil, fmt.Errorf("unknown tab %q: must be one of %s", name, strings.Join(hideableTabs, ", "))
		}
		body[name+tabAttrSuffix] = shown
	}

	req, err := newRequestWithBody(ctx, "PATCH", c.deviceConfigURL(frameID, deviceID), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create update device config request: %w", err)
	}

	var apiResp deviceConfigAPIResponse
	if err := c.patch(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update device tabs: %w", err)
	}

	cfg := apiResp.toDeviceConfig(deviceID)
	return &cfg, nil
}

func (c *Client) deviceConfigURL(frameID, deviceID string) string {
	return fmt.Sprintf("%s/frames/%s/devices/%s/device_config", c.effectiveURL(), pathSeg(frameID), pathSeg(deviceID))
}

// GetAvatars retrieves available avatars.
func (c *Client) GetAvatars(ctx context.Context) ([]Avatar, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/avatars", c.effectiveURL()))
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
func (c *Client) SetCurrentAlbum(ctx context.Context, frameID string, albumID int) error {
	body := struct {
		Frame struct {
			CurrentAlbumID int `json:"current_album_id"`
		} `json:"frame"`
	}{}
	body.Frame.CurrentAlbumID = albumID

	req, err := newRequestWithBody(ctx, "PATCH", fmt.Sprintf("%s/frames/%s", c.effectiveURL(), pathSeg(frameID)), body)
	if err != nil {
		return fmt.Errorf("failed to create set album request: %w", err)
	}

	if err := c.patch(req, nil); err != nil {
		return fmt.Errorf("failed to set current album: %w", err)
	}
	return nil
}

// UpdateFrameSettingsOptions holds optional screensaver fields for UpdateFrameSettings.
type UpdateFrameSettingsOptions struct {
	ScreensaverShowWeather *bool
	ScreensaverShowEvents  *bool
}

// UpdateFrameSettings patches frame-level settings.
func (c *Client) UpdateFrameSettings(ctx context.Context, frameID string, opts UpdateFrameSettingsOptions) error {
	inner := map[string]any{}
	if opts.ScreensaverShowWeather != nil {
		inner["screensaver_show_weather"] = *opts.ScreensaverShowWeather
	}
	if opts.ScreensaverShowEvents != nil {
		inner["screensaver_show_events"] = *opts.ScreensaverShowEvents
	}

	body := map[string]any{"frame": inner}
	req, err := newRequestWithBody(ctx, "PATCH", fmt.Sprintf("%s/frames/%s", c.effectiveURL(), pathSeg(frameID)), body)
	if err != nil {
		return fmt.Errorf("failed to create update frame request: %w", err)
	}

	if err := c.patch(req, nil); err != nil {
		return fmt.Errorf("failed to update frame settings: %w", err)
	}
	return nil
}

// ListAlbums retrieves the photo albums for a frame.
func (c *Client) ListAlbums(ctx context.Context, frameID string) ([]Album, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/albums", c.effectiveURL(), pathSeg(frameID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create list albums request: %w", err)
	}

	var apiResp albumsAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list albums: %w", err)
	}

	albums := make([]Album, len(apiResp.Data))
	for i := range apiResp.Data {
		albums[i] = apiResp.Data[i].toAlbum()
	}
	return albums, nil
}

// GetColors retrieves available colors.
func (c *Client) GetColors(ctx context.Context) ([]Color, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/colors", c.effectiveURL()))
	if err != nil {
		return nil, fmt.Errorf("failed to create get colors request: %w", err)
	}

	var apiResp colorAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to get colors: %w", err)
	}

	return apiResp.Data, nil
}
