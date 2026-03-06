package lib

import "fmt"

// ListCalendarEvents retrieves calendar events for a frame within a date range.
func (c *Client) ListCalendarEvents(frameID, startDate, endDate string) ([]CalendarEvent, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/calendar_events", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list calendar events request: %w", err)
	}

	params := map[string]string{}
	if startDate != "" {
		params["start_date"] = startDate
	}
	if endDate != "" {
		params["end_date"] = endDate
	}
	if len(params) > 0 {
		addQueryParams(req, params)
	}

	var events []CalendarEvent
	if err := c.get(req, &events); err != nil {
		return nil, fmt.Errorf("failed to list calendar events: %w", err)
	}

	return events, nil
}

// CreateCalendarEvent creates a new calendar event on a frame.
func (c *Client) CreateCalendarEvent(frameID string, event CalendarEventData) (*CalendarEvent, error) {
	reqBody := CalendarEventRequest{CalendarEvent: event}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/calendar_events", SkylightURL, frameID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar event request: %w", err)
	}

	var created CalendarEvent
	if err := c.post(req, &created); err != nil {
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}

	return &created, nil
}

// UpdateCalendarEvent updates an existing calendar event.
func (c *Client) UpdateCalendarEvent(frameID, eventID string, event CalendarEventData) (*CalendarEvent, error) {
	reqBody := CalendarEventRequest{CalendarEvent: event}

	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/calendar_events/%s", SkylightURL, frameID, eventID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create update calendar event request: %w", err)
	}

	var updated CalendarEvent
	if err := c.put(req, &updated); err != nil {
		return nil, fmt.Errorf("failed to update calendar event: %w", err)
	}

	return &updated, nil
}

// DeleteCalendarEvent deletes a calendar event.
func (c *Client) DeleteCalendarEvent(frameID, eventID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/calendar_events/%s", SkylightURL, frameID, eventID), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete calendar event request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete calendar event: %w", err)
	}

	return nil
}

// ListSourceCalendars retrieves source calendars for a frame.
func (c *Client) ListSourceCalendars(frameID string) ([]SourceCalendar, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/source_calendars", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list source calendars request: %w", err)
	}

	var calendars []SourceCalendar
	if err := c.get(req, &calendars); err != nil {
		return nil, fmt.Errorf("failed to list source calendars: %w", err)
	}

	return calendars, nil
}
