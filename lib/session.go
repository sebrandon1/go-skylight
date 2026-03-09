package lib

import "fmt"

// Login authenticates with the Skylight API using email and password.
func (c *Client) Login(email, password string) (*Session, error) {
	reqBody := SessionRequest{
		Email:    email,
		Password: password,
	}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/sessions", SkylightURL), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	var resp sessionResponse
	if err := c.post(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return &Session{
		UserID:   resp.Data.ID,
		APIToken: resp.Data.Attributes.Token,
	}, nil
}
