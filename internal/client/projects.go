package client

import (
	"context"
	"fmt"
)

func (c *Client) GetProject(ctx context.Context, id string) (*ProjectDetail, error) {
	var out ProjectDetail
	if err := c.get(ctx, "/projects/"+id, &out); err != nil {
		return nil, fmt.Errorf("get project %s: %w", id, err)
	}
	return &out, nil
}

func (c *Client) CreateProject(ctx context.Context, input ProjectInput) (*ProjectDetail, error) {
	var cr CreateResponse
	if err := c.post(ctx, "/projects", input, &cr); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return c.GetProject(ctx, cr.ID)
}

func (c *Client) UpdateProject(ctx context.Context, id string, input ProjectInput) (*ProjectDetail, error) {
	var out ProjectDetail
	if err := c.put(ctx, "/projects/"+id, input, &out); err != nil {
		return nil, fmt.Errorf("update project %s: %w", id, err)
	}
	return &out, nil
}

func (c *Client) GetProjectUsers(ctx context.Context, projectID string) ([]ProjectUser, error) {
	var out []ProjectUser
	if err := c.get(ctx, "/projects/"+projectID+"/permissions", &out); err != nil {
		return nil, fmt.Errorf("get project users %s: %w", projectID, err)
	}
	return out, nil
}

func (c *Client) SetProjectUserRole(ctx context.Context, projectID string, req SetUserProjectRoleRequest) error {
	if err := c.put(ctx, "/projects/"+projectID+"/permissions", req, nil); err != nil {
		return fmt.Errorf("set project user role: %w", err)
	}
	return nil
}
