package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func alarmsURL(base, frameID, deviceID string) string {
	return fmt.Sprintf("%s/frames/%s/devices/%s/alarms", base, pathSeg(frameID), pathSeg(deviceID))
}

func alarmURL(base, frameID, deviceID, alarmID string) string {
	return alarmsURL(base, frameID, deviceID) + "/" + pathSeg(alarmID)
}

// ListAlarms retrieves all alarms configured on a Buddy device. The API has no
// single-alarm GET endpoint, so callers needing one alarm should filter this list.
func (c *Client) ListAlarms(ctx context.Context, frameID, deviceID string) ([]Alarm, error) {
	req, err := newRequest(ctx, "GET", alarmsURL(c.effectiveURL(), frameID, deviceID))
	if err != nil {
		return nil, fmt.Errorf("failed to create list alarms request: %w", err)
	}

	var apiResp alarmAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list alarms: %w", err)
	}

	alarms := make([]Alarm, len(apiResp.Data))
	for i := range apiResp.Data {
		alarms[i] = apiResp.Data[i].toAlarm()
	}
	return alarms, nil
}

// CreateAlarm creates a new alarm on a Buddy device. It returns a nil Alarm
// without error when the API accepts the request but its response does not
// describe the created alarm.
func (c *Client) CreateAlarm(ctx context.Context, frameID, deviceID string, data AlarmData) (*Alarm, error) {
	req, err := newRequestWithBody(ctx, "POST", alarmsURL(c.effectiveURL(), frameID, deviceID), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create alarm request: %w", err)
	}

	alarm, err := c.sendAlarm(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create alarm: %w", err)
	}
	return alarm, nil
}

// UpdateAlarm updates an existing alarm on a Buddy device. Only non-nil fields
// in data are sent. If the response does not describe the alarm, it is re-read
// from ListAlarms; a nil Alarm means it could not be found there either.
func (c *Client) UpdateAlarm(ctx context.Context, frameID, deviceID, alarmID string, data AlarmData) (*Alarm, error) {
	req, err := newRequestWithBody(ctx, "PATCH", alarmURL(c.effectiveURL(), frameID, deviceID, alarmID), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create update alarm request: %w", err)
	}

	alarm, err := c.sendAlarm(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update alarm: %w", err)
	}
	if alarm != nil {
		return alarm, nil
	}

	alarms, err := c.ListAlarms(ctx, frameID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("alarm updated but re-reading it failed: %w", err)
	}
	for i := range alarms {
		if alarms[i].ID == alarmID {
			return &alarms[i], nil
		}
	}
	return nil, nil
}

// DeleteAlarm deletes an alarm from a Buddy device.
func (c *Client) DeleteAlarm(ctx context.Context, frameID, deviceID, alarmID string) error {
	req, err := newRequest(ctx, "DELETE", alarmURL(c.effectiveURL(), frameID, deviceID, alarmID))
	if err != nil {
		return fmt.Errorf("failed to create delete alarm request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete alarm: %w", err)
	}
	return nil
}

// sendAlarm performs a create/update request and parses the returned alarm.
// The response shape for alarm writes is unverified, so it accepts the JSON-API
// envelope, a flat alarm object, or an empty body (returning nil).
func (c *Client) sendAlarm(req *http.Request) (*Alarm, error) {
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if err := checkStatus(resp, body); err != nil {
		return nil, err
	}
	return parseAlarmBody(body), nil
}

func parseAlarmBody(body []byte) *Alarm {
	var single alarmAPISingleResponse
	if err := json.Unmarshal(body, &single); err == nil && single.Data.ID != "" {
		a := single.Data.toAlarm()
		return &a
	}

	var flat struct {
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(body, &flat); err != nil || len(flat.ID) == 0 || string(flat.ID) == "null" {
		return nil
	}
	var entry alarmAPIEntry
	if err := json.Unmarshal(body, &entry.Attributes); err != nil {
		return nil
	}
	entry.ID = strings.Trim(string(flat.ID), `"`)
	a := entry.toAlarm()
	return &a
}

// AlarmRRule builds a weekly alarm recurrence rule from day names (sun, mon,
// tue, wed, thu, fri, sat), e.g. FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO,WE.
// Unlike chore rules, alarm rules carry no "RRULE:" prefix.
func AlarmRRule(days []string) (string, error) {
	if len(days) == 0 {
		return "", fmt.Errorf("at least one day is required")
	}
	byDay, err := parseByDay(days)
	if err != nil {
		return "", err
	}
	return "FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=" + byDay, nil
}
