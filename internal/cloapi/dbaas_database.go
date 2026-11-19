package cloapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Database is the provider-facing view of a dbaas database. It belongs to a
// cluster and carries a per-database admin user. The admin password is
// write-only (never returned by the API), so it is not part of this struct.
type Database struct {
	ID            string
	Name          string
	ClusterID     string
	Project       string
	AdminUsername string
	BackupEnabled bool
	Status        string
	CreatedIn     string
}

func databaseFromSchema(r *gen.DbaasDatababaseSchema) Database {
	return Database{
		ID:            r.Id,
		Name:          r.Name,
		ClusterID:     r.ClusterId,
		Project:       r.Project,
		AdminUsername: r.AdminUsername,
		BackupEnabled: r.BackupEnabled,
		Status:        r.Status,
		CreatedIn:     r.CreatedIn.Format(time.RFC3339),
	}
}

// DatabaseCreateParams holds the inputs for adding a database to a cluster.
type DatabaseCreateParams struct {
	Name          string
	AdminUsername string
	AdminPassword string
}

// CreateDatabase adds a database to the cluster and returns its ID.
func (c *Client) CreateDatabase(ctx context.Context, clusterID string, p DatabaseCreateParams) (string, error) {
	body := gen.ClusterAddDatabaseJSONRequestBody{
		Name:          p.Name,
		AdminUsername: p.AdminUsername,
		AdminPassword: p.AdminPassword,
	}
	resp, err := c.gen.ClusterAddDatabaseWithResponse(ctx, clusterID, body)
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty dbaas database create response")
	}
	return resp.OK.Result.Id, nil
}

// GetDatabase returns the database's current detail.
func (c *Client) GetDatabase(ctx context.Context, id string) (*Database, error) {
	resp, err := c.gen.DbaasDatabaseDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, errors.New("cloapi: empty dbaas database detail response")
	}
	db := databaseFromSchema(resp.OK.Result)
	return &db, nil
}

// ListDatabasesByCluster returns the databases in a cluster.
func (c *Client) ListDatabasesByCluster(ctx context.Context, clusterID string, options *ListOptions) ([]Database, error) {
	return CollectPages(ctx, options, func(query map[string]string) ([]Database, int, error) {
		resp, err := c.gen.ClusterDbaasDatabasesListWithResponse(ctx, clusterID, gen.WithQuery(query))
		if err != nil {
			return nil, 0, err
		}
		if resp.OK == nil {
			return nil, 0, fmt.Errorf("missing ListDatabasesByCluster response")
		}
		if resp.OK.Result == nil {
			return nil, resp.OK.Count, nil
		}
		return databasesFromList(resp.OK.Result), resp.OK.Count, nil
	})
}

// ListDatabasesByProject returns all dbaas databases in a project.
func (c *Client) ListDatabasesByProject(ctx context.Context, projectID string, options *ListOptions) ([]Database, error) {
	return CollectPages(ctx, options, func(query map[string]string) ([]Database, int, error) {
		resp, err := c.gen.ProjectDbaasDatabasesListWithResponse(ctx, projectID, gen.WithQuery(query))
		if err != nil {
			return nil, 0, err
		}
		if resp.OK == nil {
			return nil, 0, fmt.Errorf("missing ListDatabasesByProject response")
		}
		if resp.OK.Result == nil {
			return nil, resp.OK.Count, nil
		}
		return databasesFromList(resp.OK.Result), resp.OK.Count, nil
	})
}

func databasesFromList(items *[]gen.DbaasDatababaseSchema) []Database {
	if items == nil {
		return nil
	}
	list := *items
	out := make([]Database, 0, len(list))
	for i := range list {
		out = append(out, databaseFromSchema(&list[i]))
	}
	return out
}

// RestoreAdminPassword rotates the database's admin password.
func (c *Client) RestoreAdminPassword(ctx context.Context, id, password string) error {
	_, err := c.gen.DbaasRestoreAdminPasswordWithResponse(ctx, id, gen.DbaasRestoreAdminPasswordJSONRequestBody{Password: password})
	return err
}

// DeleteDatabase deletes a database.
func (c *Client) DeleteDatabase(ctx context.Context, id string) error {
	_, err := c.gen.DbaasDatabaseDeleteWithResponse(ctx, id)
	return err
}
