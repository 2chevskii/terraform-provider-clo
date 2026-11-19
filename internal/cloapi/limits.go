package cloapi

import (
	"context"
	gen "github.com/clo-ru/cloapi-go-client/v3"
)

func limitField[T any](values map[string]T, key string) *T {
	v, ok := values[key]
	if !ok {
		return nil
	}
	return &v
}
func (c *Client) UpdateProjectLimits(ctx context.Context, project string, values map[string]int) error {
	_, err := c.gen.ProjectPatchLimitsWithResponse(ctx, project, gen.ProjectPatchLimitsJSONRequestBody{
		Servers: limitField(values, "servers"), DiskSpace: limitField(values, "disk-space"), Loadbalancer: limitField(values, "loadbalancer"),
		S3MaxSpace: limitField(values, "s3-max-space"), Addresses: limitField(values, "addresses"), DbaasClusters: limitField(values, "dbaas-clusters"),
	})
	return err
}
func (c *Client) UpdateAccountLimits(ctx context.Context, values map[string]map[string]int) error {
	_, err := c.gen.AccountPatchLimitsWithResponse(ctx, gen.AccountPatchLimitsJSONRequestBody{
		Servers: limitField(values, "servers"), DiskSpace: limitField(values, "disk-space"), Loadbalancer: limitField(values, "loadbalancer"),
		S3MaxSpace: limitField(values, "s3-max-space"), Addresses: limitField(values, "addresses"), DbaasClusters: limitField(values, "dbaas-clusters"),
	})
	return err
}
