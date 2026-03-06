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

	var session Session
	if err := c.post(req, &session); err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return &session, nil
}
