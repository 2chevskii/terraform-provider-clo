package cloapi

import (
	"context"
	"fmt"
	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Image is the provider-facing view of an OS image.
type Image struct {
	ID             string
	Name           string
	OSDistribution string
	OSFamily       string
	OSVersion      string
}

// ListImages returns the project's OS images.
func (c *Client) ListImages(ctx context.Context, projectID string, options *ListOptions) ([]Image, error) {
	return CollectPages(ctx, options, func(query map[string]string) ([]Image, int, error) {
		resp, err := c.gen.ProjectImagesListWithResponse(ctx, projectID, gen.WithQuery(query))
		if err != nil {
			return nil, 0, err
		}
		if resp.OK == nil {
			return nil, 0, fmt.Errorf("missing ListImages response")
		}
		if resp.OK.Result == nil {
			return nil, resp.OK.Count, nil
		}
		items := *resp.OK.Result
		out := make([]Image, 0, len(items))
		for _, im := range items {
			out = append(out, Image{
				ID:             im.Id,
				Name:           im.Name,
				OSDistribution: im.OperationSystem.Distribution,
				OSFamily:       im.OperationSystem.OsFamily,
				OSVersion:      im.OperationSystem.Version,
			})
		}
		return out, resp.OK.Count, nil
	})
}
