package client

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) GetProcess(ctx context.Context, processID string) (*ProcessDetail, error) {
	var out ProcessDetail
	if err := c.get(ctx, "/processes/"+url.PathEscape(processID), &out); err != nil {
		return nil, fmt.Errorf("get process %s: %w", processID, err)
	}
	return &out, nil
}

func (c *Client) CreateProcess(ctx context.Context, req ProcessInput) (string, error) {
	var out CreateResponse
	if err := c.post(ctx, "/processes", req, &out); err != nil {
		return "", fmt.Errorf("create process: %w", err)
	}
	return out.ID, nil
}

func (c *Client) UpdateProcess(ctx context.Context, processID string, req ProcessInput) error {
	if err := c.put(ctx, "/processes/"+url.PathEscape(processID), req, nil); err != nil {
		return fmt.Errorf("update process %s: %w", processID, err)
	}
	return nil
}

func (c *Client) ArchiveProcess(ctx context.Context, processID string) error {
	if err := c.delete(ctx, "/processes/"+url.PathEscape(processID)); err != nil {
		return fmt.Errorf("archive process %s: %w", processID, err)
	}
	return nil
}
