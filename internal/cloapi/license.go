package cloapi

import (
	"context"
	"errors"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

func (c *Client) AddServerLicense(ctx context.Context, server, addon, name string, value *int) (string, error) {
	resp, err := c.gen.ServerAddLicenseWithResponse(ctx, server, gen.ServerAddLicenseJSONRequestBody{Addon: addon, Name: name, Value: value})
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil || resp.OK.Result.Id == "" {
		return "", errors.New("CLO returned no license ID")
	}
	return resp.OK.Result.Id, nil
}
func (c *Client) UpdateLicense(ctx context.Context, id, name string, value *int) error {
	_, err := c.gen.LicenseUpdateWithResponse(ctx, id, gen.LicenseUpdateJSONRequestBody{Name: name, Value: value})
	return err
}
func (c *Client) DeleteLicense(ctx context.Context, id string) error {
	_, err := c.gen.LicenseDeleteWithResponse(ctx, id)
	return err
}
func (c *Client) RebootServer(ctx context.Context, id string) error {
	_, err := c.gen.ServerRebootWithResponse(ctx, id)
	return err
}
func (c *Client) RescueServer(ctx context.Context, id string) error {
	_, err := c.gen.ServerRescueWithResponse(ctx, id)
	return err
}
