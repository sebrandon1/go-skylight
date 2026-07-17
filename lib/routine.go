package lib

import "fmt"

// Routine represents a Skylight routine (ordered recurring task list).
type Routine struct {
	ID         string        `json:"id,omitempty"`
	Title      string        `json:"title,omitempty"`
	AssigneeID string        `json:"assignee_id,omitempty"`
	Steps      []RoutineStep `json:"steps,omitempty"`
	CreatedAt  string        `json:"created_at,omitempty"`
	UpdatedAt  string        `json:"updated_at,omitempty"`
}

// RoutineStep represents a single step within a routine.
type RoutineStep struct {
	ID       string `json:"id,omitempty"`
	Title    string `json:"title,omitempty"`
	Position int    `json:"position,omitempty"`
}

// RoutineData holds the fields for create/update routine requests.
type RoutineData struct {
	Title      string   `json:"title,omitempty"`
	AssigneeID string   `json:"assignee_id,omitempty"`
	Steps      []string `json:"steps,omitempty"`
}

type reorderRoutinesRequest struct {
	IDs []string `json:"ids"`
}

type routineAPIResponse struct {
	Data []routineAPIEntry `json:"data"`
}

type routineAPISingleResponse struct {
	Data routineAPIEntry `json:"data"`
}

type routineAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Title      string        `json:"title"`
		AssigneeID string        `json:"assignee_id"`
		Steps      []RoutineStep `json:"steps"`
		CreatedAt  string        `json:"created_at"`
		UpdatedAt  string        `json:"updated_at"`
	} `json:"attributes"`
}

func (e *routineAPIEntry) toRoutine() Routine {
	return Routine{
		ID:         e.ID,
		Title:      e.Attributes.Title,
		AssigneeID: e.Attributes.AssigneeID,
		Steps:      e.Attributes.Steps,
		CreatedAt:  e.Attributes.CreatedAt,
		UpdatedAt:  e.Attributes.UpdatedAt,
	}
}

func (c *Client) ListRoutines(frameID string) ([]Routine, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/routines", c.effectiveURL(), pathSeg(frameID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create list routines request: %w", err)
	}

	var apiResp routineAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list routines: %w", err)
	}

	routines := make([]Routine, len(apiResp.Data))
	for i := range apiResp.Data {
		routines[i] = apiResp.Data[i].toRoutine()
	}
	return routines, nil
}

func (c *Client) CreateRoutine(frameID string, data RoutineData) (*Routine, error) {
	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/routines", c.effectiveURL(), pathSeg(frameID)), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create routine request: %w", err)
	}

	var apiResp routineAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create routine: %w", err)
	}

	result := apiResp.Data.toRoutine()
	return &result, nil
}

func (c *Client) UpdateRoutine(frameID, routineID string, data RoutineData) (*Routine, error) {
	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/routines/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(routineID)), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create update routine request: %w", err)
	}

	var apiResp routineAPISingleResponse
	if err := c.put(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update routine: %w", err)
	}

	result := apiResp.Data.toRoutine()
	return &result, nil
}

func (c *Client) DeleteRoutine(frameID, routineID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/routines/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(routineID)))
	if err != nil {
		return fmt.Errorf("failed to create delete routine request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete routine: %w", err)
	}

	return nil
}

func (c *Client) ReorderRoutines(frameID string, routineIDs []string) error {
	body := reorderRoutinesRequest{IDs: routineIDs}

	req, err := newRequestWithBody("PATCH", fmt.Sprintf("%s/frames/%s/routines/reorder", c.effectiveURL(), pathSeg(frameID)), body)
	if err != nil {
		return fmt.Errorf("failed to create reorder routines request: %w", err)
	}

	var result any
	if err := c.patch(req, &result); err != nil {
		return fmt.Errorf("failed to reorder routines: %w", err)
	}

	return nil
}
