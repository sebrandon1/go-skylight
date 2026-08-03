package lib

import (
	"context"
	"fmt"
)

// ListCalendarEvents retrieves calendar events for a frame within a date range.
func (c *Client) ListCalendarEvents(ctx context.Context, frameID, dateMin, dateMax, timezone string) ([]CalendarEvent, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/calendar_events", c.effectiveURL(), pathSeg(frameID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create list calendar events request: %w", err)
	}

	params := map[string]string{}
	if dateMin != "" {
		params["date_min"] = dateMin
	}
	if dateMax != "" {
		params["date_max"] = dateMax
	}
	if timezone != "" {
		params["timezone"] = timezone
	}
	if len(params) > 0 {
		addQueryParams(req, params)
	}

	var apiResp calendarAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list calendar events: %w", err)
	}

	events := make([]CalendarEvent, len(apiResp.Data))
	for i := range apiResp.Data {
		events[i] = apiResp.Data[i].toCalendarEvent()
	}
	return events, nil
}

// CreateCalendarEvent creates a new calendar event on a frame.
func (c *Client) CreateCalendarEvent(ctx context.Context, frameID string, event CalendarEventData) (*CalendarEvent, error) {
	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/calendar_events", c.effectiveURL(), pathSeg(frameID)), event)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar event request: %w", err)
	}

	var apiResp calendarAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}

	result := apiResp.Data.toCalendarEvent()
	return &result, nil
}

// UpdateCalendarEvent updates an existing calendar event.
func (c *Client) UpdateCalendarEvent(ctx context.Context, frameID, eventID string, event CalendarEventData) (*CalendarEvent, error) {
	req, err := newRequestWithBody(ctx, "PUT", fmt.Sprintf("%s/frames/%s/calendar_events/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(eventID)), event)
	if err != nil {
		return nil, fmt.Errorf("failed to create update calendar event request: %w", err)
	}

	var apiResp calendarAPISingleResponse
	if err := c.put(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update calendar event: %w", err)
	}

	result := apiResp.Data.toCalendarEvent()
	return &result, nil
}

// DeleteCalendarEvent deletes a calendar event.
func (c *Client) DeleteCalendarEvent(ctx context.Context, frameID, eventID string) error {
	req, err := newRequest(ctx, "DELETE", fmt.Sprintf("%s/frames/%s/calendar_events/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(eventID)))
	if err != nil {
		return fmt.Errorf("failed to create delete calendar event request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete calendar event: %w", err)
	}

	return nil
}

// ListSourceCalendars retrieves source calendars for a frame.
func (c *Client) ListSourceCalendars(ctx context.Context, frameID string) ([]SourceCalendar, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/source_calendars", c.effectiveURL(), pathSeg(frameID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create list source calendars request: %w", err)
	}

	var apiResp sourceCalendarAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list source calendars: %w", err)
	}

	calendars := make([]SourceCalendar, len(apiResp.Data))
	for i := range apiResp.Data {
		calendars[i] = apiResp.Data[i].toSourceCalendar()
	}
	return calendars, nil
}
