package lib

import (
	"context"
	"fmt"
	"time"
)

// ListMessages retrieves household messages delivering between after and
// before (YYYY-MM-DD, both required by the API). Recurring messages appear
// once per occurrence, all with the same ID.
func (c *Client) ListMessages(ctx context.Context, frameID, after, before string) ([]Message, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/nudges", c.effectiveURL(), pathSeg(frameID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create list messages request: %w", err)
	}
	addQueryParams(req, map[string]string{"after": after, "before": before})

	var apiResp messageAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	messages := make([]Message, len(apiResp.Data))
	for i := range apiResp.Data {
		messages[i] = apiResp.Data[i].toMessage()
	}
	return messages, nil
}

// SendMessage creates a household message. The API requires a delivery time,
// so an empty data.DeliverAt is set to now, which delivers it immediately.
func (c *Client) SendMessage(ctx context.Context, frameID string, data MessageData) (*Message, error) {
	if data.DeliverAt == "" {
		data.DeliverAt = time.Now().UTC().Format(time.RFC3339)
	}
	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/nudges", c.effectiveURL(), pathSeg(frameID)), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create send message request: %w", err)
	}

	var apiResp messageAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	result := apiResp.Data.toMessage()
	return &result, nil
}

// UpdateMessage updates a household message. Only non-zero fields
// in data are sent; CategoryIDs, when set, replaces the recipient list.
func (c *Client) UpdateMessage(ctx context.Context, frameID, messageID string, data MessageData) (*Message, error) {
	req, err := newRequestWithBody(ctx, "PATCH", fmt.Sprintf("%s/frames/%s/nudges/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(messageID)), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create update message request: %w", err)
	}

	var apiResp messageAPISingleResponse
	if err := c.patch(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	result := apiResp.Data.toMessage()
	return &result, nil
}

// DeleteMessage deletes a household message.
func (c *Client) DeleteMessage(ctx context.Context, frameID, messageID string) error {
	req, err := newRequest(ctx, "DELETE", fmt.Sprintf("%s/frames/%s/nudges/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(messageID)))
	if err != nil {
		return fmt.Errorf("failed to create delete message request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}
