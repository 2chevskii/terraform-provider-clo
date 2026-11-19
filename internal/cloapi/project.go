package cloapi

import (
	"context"
	"errors"
	gen "github.com/clo-ru/cloapi-go-client/v3"
	"time"
)

// Project is the provider-facing view of a project.
type Project struct {
	ID             string
	Name           string
	DisplayName    string
	Description    string
	SwitchStatus   string
	Status         string
	CreatedIn      string
	StoppingReason string
	HasAbuse       bool
}

// ListProjects returns the account's projects (single page, matching v2 behavior).
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	resp, err := c.gen.ProjectListWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Project, 0, len(items))
	for _, p := range items {
		out = append(out, projectFromSchema(p))
	}
	return out, nil
}

func projectFromSchema(p gen.ProjectDetailSchema) Project {
	project := Project{ID: p.Id, Name: p.Name, Status: p.Status, CreatedIn: p.Created.Format(time.RFC3339), SwitchStatus: p.SwitchStatus}
	if p.HasAbuse != nil {
		project.HasAbuse = *p.HasAbuse
	}
	if p.StoppingReason != nil {
		project.StoppingReason = *p.StoppingReason
	}
	if p.DisplayName != nil {
		project.DisplayName = *p.DisplayName
	}
	if p.Description != nil {
		project.Description = *p.Description
	}
	return project
}
func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	r, err := c.gen.ProjectDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.OK == nil || r.OK.Result == nil {
		return nil, errors.New("missing project detail")
	}
	project := projectFromSchema(*r.OK.Result)
	return &project, nil
}
func (c *Client) CreateProject(ctx context.Context, displayName, description *string, limitName *int) (string, error) {
	r, err := c.gen.ProjectCreateWithResponse(ctx, gen.ProjectCreateJSONRequestBody{DisplayName: displayName, Description: description, LimitName: limitName})
	if err != nil {
		return "", err
	}
	if r.OK == nil || r.OK.Result == nil || r.OK.Result.Id == "" {
		return "", errors.New("missing created project ID")
	}
	return r.OK.Result.Id, nil
}
func (c *Client) UpdateProject(ctx context.Context, id string, displayName, description *string) error {
	_, err := c.gen.ProjectPatchDisplayNameDescriptionWithResponse(ctx, id, gen.ProjectPatchDisplayNameDescriptionJSONRequestBody{DisplayName: displayName, Description: description})
	return err
}
func (c *Client) SetProjectEnabled(ctx context.Context, id string, enabled bool) error {
	if enabled {
		_, err := c.gen.ProjectStartWithResponse(ctx, id)
		return err
	}
	_, err := c.gen.ProjectStopWithResponse(ctx, id)
	return err
}
func (c *Client) DeleteProject(ctx context.Context, id string) error {
	_, err := c.gen.ProjectDeleteWithResponse(ctx, id)
	return err
}
