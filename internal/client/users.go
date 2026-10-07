package client

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) GetUser(ctx context.Context, username string) (*UserDetail, error) {
	var out UserDetail
	if err := c.get(ctx, "/users/"+url.PathEscape(username), &out); err != nil {
		return nil, fmt.Errorf("get user %s: %w", username, err)
	}
	return &out, nil
}

func (c *Client) InviteUser(ctx context.Context, req InviteUserRequest) error {
	var out InviteUserResponse
	if err := c.post(ctx, "/users", req, &out); err != nil {
		return fmt.Errorf("invite user: %w", err)
	}
	return nil
}

func (c *Client) UpdateUser(ctx context.Context, username string, req UpdateUserRequest) (*UserDetail, error) {
	if err := c.put(ctx, "/users/"+url.PathEscape(username), req, nil); err != nil {
		return nil, fmt.Errorf("update user %s: %w", username, err)
	}
	return c.GetUser(ctx, username)
}
