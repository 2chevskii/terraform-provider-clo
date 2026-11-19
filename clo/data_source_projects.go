package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceProjects() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of the projects",
		ReadContext: dataSourceProjectsRead,
		Schema: map[string]*schema.Schema{
			"query":       listQuerySchema(),
			"total_count": computedField(schema.TypeInt),
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the project",
							Type:        schema.TypeString, Computed: true},
						"name": {
							Description: "Name of the project",
							Type:        schema.TypeString, Computed: true},
						"switch_status": computedField(schema.TypeString),
						"display_name":  computedField(schema.TypeString),
						"description":   computedField(schema.TypeString),
						"status":        {Type: schema.TypeString, Computed: true},
						"created_in": {
							Description: "Timestamp the project was created",
							Type:        schema.TypeString, Computed: true},
						"stopping_reason": {
							Description: "A reason the project was stopped",
							Type:        schema.TypeString, Computed: true},
						"has_abuse": {
							Description: "Is the project has an abuse issues",
							Type:        schema.TypeBool, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceProjectsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	options := listOptions(d)
	projects, err := cli.ListProjects(ctx, options)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("total_count", options.Total); err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("result", flattenProjectResults(projects)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenProjectResults(pr []cloapi.Project) []interface{} {
	res := make([]interface{}, 0, len(pr))
	for _, p := range pr {
		res = append(res, map[string]interface{}{
			"id":     p.ID,
			"name":   p.Name,
			"status": p.Status, "switch_status": p.SwitchStatus, "display_name": p.DisplayName, "description": p.Description,
			"has_abuse":       p.HasAbuse,
			"created_in":      p.CreatedIn,
			"stopping_reason": p.StoppingReason,
		})
	}
	return res
}
