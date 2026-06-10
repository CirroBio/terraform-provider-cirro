package client

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) ListAgents(ctx context.Context) ([]AgentDetail, error) {
	var out []AgentDetail
	if err := c.get(ctx, "/agents", &out); err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	return out, nil
}

func (c *Client) GetAgent(ctx context.Context, id string) (*AgentDetail, error) {
	agents, err := c.ListAgents(ctx)
	if err != nil {
		return nil, err
	}
	for i := range agents {
		if agents[i].ID == id {
			return &agents[i], nil
		}
	}
	return nil, fmt.Errorf("agent %q not found", id)
}

func (c *Client) CreateAgent(ctx context.Context, input AgentInput) (*AgentDetail, error) {
	var cr CreateResponse
	if err := c.post(ctx, "/agents", input, &cr); err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}
	return c.GetAgent(ctx, cr.ID)
}

func (c *Client) UpdateAgent(ctx context.Context, id string, input AgentInput) error {
	if err := c.put(ctx, "/agents/"+id, input, nil); err != nil {
		return fmt.Errorf("update agent %s: %w", id, err)
	}
	return nil
}

func (c *Client) DeleteAgent(ctx context.Context, id string) error {
	if err := c.do(ctx, http.MethodDelete, "/agents/"+id, nil, nil); err != nil {
		return fmt.Errorf("delete agent %s: %w", id, err)
	}
	return nil
}
