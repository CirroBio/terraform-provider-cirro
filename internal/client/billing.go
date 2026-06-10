package client

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) ListBillingAccounts(ctx context.Context) ([]BillingAccount, error) {
	var out []BillingAccount
	if err := c.get(ctx, "/billing", &out); err != nil {
		return nil, fmt.Errorf("list billing accounts: %w", err)
	}
	return out, nil
}

func (c *Client) GetBillingAccount(ctx context.Context, id string) (*BillingAccount, error) {
	accounts, err := c.ListBillingAccounts(ctx)
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		if accounts[i].ID == id {
			return &accounts[i], nil
		}
	}
	return nil, fmt.Errorf("billing account %q not found", id)
}

func (c *Client) CreateBillingAccount(ctx context.Context, req BillingAccountRequest) (*BillingAccount, error) {
	var cr CreateResponse
	if err := c.post(ctx, "/billing", req, &cr); err != nil {
		return nil, fmt.Errorf("create billing account: %w", err)
	}
	return c.GetBillingAccount(ctx, cr.ID)
}

func (c *Client) UpdateBillingAccount(ctx context.Context, id string, req BillingAccountRequest) error {
	if err := c.put(ctx, "/billing/"+id, req, nil); err != nil {
		return fmt.Errorf("update billing account %s: %w", id, err)
	}
	return nil
}

func (c *Client) DeleteBillingAccount(ctx context.Context, id string) error {
	if err := c.do(ctx, http.MethodDelete, "/billing/"+id, nil, nil); err != nil {
		return fmt.Errorf("delete billing account %s: %w", id, err)
	}
	return nil
}
