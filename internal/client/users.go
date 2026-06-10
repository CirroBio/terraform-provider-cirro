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

// FindUserByEmail searches the user list for a user with the given email address.
// Cirro's list endpoint filters by username pattern, so we do a broad search and
// filter client-side by email.
func (c *Client) FindUserByEmail(ctx context.Context, email string) (*UserDto, error) {
	var resp PaginatedUsersResponse
	if err := c.get(ctx, "/users", &resp); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	for i := range resp.Items {
		if resp.Items[i].Email == email {
			return &resp.Items[i], nil
		}
	}

	// Paginate if needed
	for resp.NextToken != "" {
		if err := c.get(ctx, "/users?nextToken="+url.QueryEscape(resp.NextToken), &resp); err != nil {
			return nil, fmt.Errorf("list users (paginated): %w", err)
		}
		for i := range resp.Items {
			if resp.Items[i].Email == email {
				return &resp.Items[i], nil
			}
		}
	}

	return nil, fmt.Errorf("user with email %q not found", email)
}

func (c *Client) InviteUser(ctx context.Context, req InviteUserRequest) error {
	var out InviteUserResponse
	if err := c.post(ctx, "/users", req, &out); err != nil {
		return fmt.Errorf("invite user: %w", err)
	}
	return nil
}

func (c *Client) UpdateUser(ctx context.Context, username string, req UpdateUserRequest) (*UserDetail, error) {
	var out UserDetail
	if err := c.put(ctx, "/users/"+url.PathEscape(username), req, &out); err != nil {
		return nil, fmt.Errorf("update user %s: %w", username, err)
	}
	return &out, nil
}
