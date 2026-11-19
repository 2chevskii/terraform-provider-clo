package clo

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var writableLimits = []string{"addresses", "dbaas_clusters", "disk_space", "loadbalancer", "s3_max_space", "servers"}

// Both APIs manage project limits. The account endpoint accepts a map of project
// IDs per limit. The project endpoint scopes the same values via its path.
func resourceLimits(account bool) *schema.Resource {
	fields := map[string]*schema.Schema{}
	if !account {
		fields["project_id"] = &schema.Schema{Type: schema.TypeString, Required: true, ForceNew: true}
		fields["limits"] = &schema.Schema{Type: schema.TypeMap, Required: true, Elem: &schema.Schema{Type: schema.TypeInt}, Description: "Maximum values keyed by addresses, dbaas-clusters, disk-space, loadbalancer, s3-max-space or servers."}
	} else {
		for _, key := range writableLimits {
			fields[key] = &schema.Schema{Type: schema.TypeMap, Optional: true, AtLeastOneOf: writableLimits, Elem: &schema.Schema{Type: schema.TypeInt}, Description: "Map of project IDs to maximum values for " + strings.ReplaceAll(key, "_", " ") + "."}
		}
	}
	// Preserve omitted limits in CLO. There is no documented reset operation.
	desc := "Manage the specified project limits. Removing a limit or destroying this resource leaves its last value in CLO. Manage each project and limit pair in one resource."
	read := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		managed := limitValues(d, account)
		actual, err := readLimitValues(ctx, m.(*providerMeta).v3, d.Id(), account, managed)
		if !account && cloapi.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		if err != nil {
			return diag.FromErr(err)
		}
		if account {
			for name := range managed {
				if err := d.Set(strings.ReplaceAll(name, "-", "_"), actual[name]); err != nil {
					return diag.FromErr(err)
				}
			}
		} else {
			limits := map[string]int{}
			for name, projects := range actual {
				limits[name] = projects[d.Id()]
			}
			if err := d.Set("limits", limits); err != nil {
				return diag.FromErr(err)
			}
		}
		return nil
	}
	apply := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		cli := m.(*providerMeta).v3
		id := "account"
		values := limitValues(d, account)
		if len(values) == 0 {
			return diag.Errorf("specify at least one project limit")
		}
		var err error
		if account {
			for name, projects := range values {
				for project, n := range projects {
					if project == "" || n < 0 {
						return diag.Errorf("invalid %s limit for project %q", name, project)
					}
				}
			}
			err = cli.UpdateAccountLimits(ctx, values)
		} else {
			id = d.Get("project_id").(string)
			patch := map[string]int{}
			for name, projects := range values {
				if !slices.Contains(writableLimits, strings.ReplaceAll(name, "-", "_")) || strings.Contains(name, "_") || projects[id] < 0 {
					return diag.Errorf("invalid project limit %s=%d", name, projects[id])
				}
				patch[name] = projects[id]
			}
			err = cli.UpdateProjectLimits(ctx, id, patch)
		}
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(id)
		err = waitForState(ctx, d.Timeout(schema.TimeoutUpdate), []string{"pending"}, []string{"ready"}, func() (interface{}, string, error) {
			actual, err := readLimitValues(ctx, cli, id, account, values)
			if err != nil {
				return nil, "", err
			}
			for key, expected := range values {
				if !maps.Equal(expected, actual[key]) {
					return actual, "pending", nil
				}
			}
			return actual, "ready", nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
	return &schema.Resource{Description: desc, Schema: fields, CreateContext: apply, UpdateContext: apply, ReadContext: read, DeleteContext: schema.NoopContext, Timeouts: &schema.ResourceTimeout{Default: schema.DefaultTimeout(10 * time.Minute)}}
}

func limitValues(d *schema.ResourceData, account bool) map[string]map[string]int {
	values := map[string]map[string]int{}
	if !account {
		for name, n := range d.Get("limits").(map[string]interface{}) {
			values[name] = map[string]int{d.Get("project_id").(string): n.(int)}
		}
		return values
	}
	for _, key := range writableLimits {
		projects := map[string]int{}
		for project, n := range d.Get(key).(map[string]interface{}) {
			projects[project] = n.(int)
		}
		if len(projects) > 0 {
			values[strings.ReplaceAll(key, "_", "-")] = projects
		}
	}
	return values
}

func readLimitValues(ctx context.Context, cli *cloapi.Client, id string, account bool, managed map[string]map[string]int) (map[string]map[string]int, error) {
	op := "ProjectLimitsList"
	if account {
		op = "AccountProjectsLimits"
	}
	rows, _, err := readQuery(ctx, cli, op, id, map[string]string{}, true, true)
	if err != nil {
		return nil, err
	}
	actual := map[string]map[string]int{}
	for _, row := range rows {
		name, _ := row["name"].(string)
		project := id
		if account {
			project, _ = row["project_id"].(string)
		}
		if _, ok := managed[name][project]; !ok || row["max"] == nil || row["enabled"] == false {
			continue
		}
		number, ok := row["max"].(json.Number)
		if !ok {
			return nil, fmt.Errorf("invalid maximum for %s", name)
		}
		max, err := number.Int64()
		if err != nil {
			return nil, err
		}
		if actual[name] == nil {
			actual[name] = map[string]int{}
		}
		actual[name][project] = int(max)
	}
	return actual, nil
}
