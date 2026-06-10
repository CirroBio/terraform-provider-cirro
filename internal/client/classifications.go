package client

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) GetClassification(ctx context.Context, id string) (*GovernanceClassification, error) {
	var out GovernanceClassification
	if err := c.get(ctx, "/governance/classifications/"+id, &out); err != nil {
		return nil, fmt.Errorf("get classification %s: %w", id, err)
	}
	return &out, nil
}

func (c *Client) CreateClassification(ctx context.Context, input ClassificationInput) (*GovernanceClassification, error) {
	var out GovernanceClassification
	if err := c.post(ctx, "/governance/classifications", input, &out); err != nil {
		return nil, fmt.Errorf("create classification: %w", err)
	}
	return &out, nil
}

func (c *Client) UpdateClassification(ctx context.Context, id string, input ClassificationInput) (*GovernanceClassification, error) {
	var out GovernanceClassification
	if err := c.put(ctx, "/governance/classifications/"+id, input, &out); err != nil {
		return nil, fmt.Errorf("update classification %s: %w", id, err)
	}
	return &out, nil
}

func (c *Client) DeleteClassification(ctx context.Context, id string) error {
	if err := c.do(ctx, http.MethodDelete, "/governance/classifications/"+id, nil, nil); err != nil {
		return fmt.Errorf("delete classification %s: %w", id, err)
	}
	return nil
}
