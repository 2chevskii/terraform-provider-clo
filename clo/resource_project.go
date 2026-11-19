package clo

import (
	"context"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage a CLO project, its metadata and its power state. Import using the project ID.",
		CreateContext: projectCreate, ReadContext: projectRead, UpdateContext: projectUpdate, DeleteContext: projectDelete,
		Importer: &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: &schema.ResourceTimeout{Default: schema.DefaultTimeout(20 * time.Minute)},
		Schema: map[string]*schema.Schema{
			"name":         computedField(schema.TypeString),
			"display_name": {Type: schema.TypeString, Optional: true, Computed: true, ValidateFunc: validation.StringLenBetween(1, 40), Description: "Project display name."},
			"description":  {Type: schema.TypeString, Optional: true, Computed: true, ValidateFunc: validation.StringLenBetween(1, 300), Description: "Project description."},
			"limit_name":   {Type: schema.TypeInt, Optional: true, ForceNew: true, ValidateFunc: validation.IntAtLeast(0), Description: "Optional limit selector used only during project creation."},
			"enabled":      {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Desired project power state."},
			"status":       computedField(schema.TypeString), "created_in": computedField(schema.TypeString), "stopping_reason": computedField(schema.TypeString), "has_abuse": computedField(schema.TypeBool),
		}}
}

func projectString(d *schema.ResourceData, key string) *string {
	if v, ok := d.GetOk(key); ok {
		s := v.(string)
		return &s
	}
	return nil
}

func projectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	var limit *int
	raw := d.GetRawConfig().GetAttr("limit_name")
	if raw.IsKnown() && !raw.IsNull() {
		v := d.Get("limit_name").(int)
		limit = &v
	}
	id, err := cli.CreateProject(ctx, projectString(d, "display_name"), projectString(d, "description"), limit)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	if err := waitProject(ctx, d, cli, schema.TimeoutCreate, func(p *cloapi.Project) bool { return p.Status == "ACTIVE" || p.Status == "STOPPED" }); err != nil {
		return diag.FromErr(err)
	}
	enabled := d.GetRawConfig().GetAttr("enabled")
	if enabled.IsKnown() && !enabled.IsNull() {
		if err := projectPower(ctx, d, cli, enabled.True(), schema.TimeoutCreate); err != nil {
			return diag.FromErr(err)
		}
	}
	return projectRead(ctx, d, m)
}

func projectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	p, err := m.(*providerMeta).v3.GetProject(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if p.Status == "DELETED" {
		d.SetId("")
		return nil
	}
	for key, value := range map[string]interface{}{"name": p.Name, "description": p.Description, "display_name": p.DisplayName, "enabled": p.SwitchStatus == "ON", "status": p.Status, "created_in": p.CreatedIn, "stopping_reason": p.StoppingReason, "has_abuse": p.HasAbuse} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func projectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	var name, desc *string
	if d.HasChange("display_name") {
		name = projectString(d, "display_name")
	}
	if d.HasChange("description") {
		desc = projectString(d, "description")
	}
	if name != nil || desc != nil {
		if err := cli.UpdateProject(ctx, d.Id(), name, desc); err != nil {
			return diag.FromErr(err)
		}
		if err := waitProject(ctx, d, cli, schema.TimeoutUpdate, func(p *cloapi.Project) bool {
			return (name == nil || *name == p.DisplayName) && (desc == nil || *desc == p.Description)
		}); err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("enabled") {
		if err := projectPower(ctx, d, cli, d.Get("enabled").(bool), schema.TimeoutUpdate); err != nil {
			return diag.FromErr(err)
		}
	}
	return projectRead(ctx, d, m)
}
func projectPower(ctx context.Context, d *schema.ResourceData, cli *cloapi.Client, enabled bool, timeout string) error {
	p, err := cli.GetProject(ctx, d.Id())
	if err != nil {
		return err
	}
	if (p.SwitchStatus == "ON") != enabled {
		if err := cli.SetProjectEnabled(ctx, d.Id(), enabled); err != nil {
			return err
		}
	}
	target := "STOPPED"
	if enabled {
		target = "ACTIVE"
	}
	if p.Status == target && (p.SwitchStatus == "ON") == enabled {
		return nil
	}
	return waitProject(ctx, d, cli, timeout, func(p *cloapi.Project) bool { return p.Status == target && (p.SwitchStatus == "ON") == enabled })
}
func waitProject(ctx context.Context, d *schema.ResourceData, cli *cloapi.Client, timeout string, ready func(*cloapi.Project) bool) error {
	return waitForState(ctx, d.Timeout(timeout), []string{"pending"}, []string{"ready"}, func() (interface{}, string, error) {
		p, err := cli.GetProject(ctx, d.Id())
		if cloapi.IsNotFound(err) {
			switch timeout {
			case schema.TimeoutCreate:
				return struct{}{}, "pending", nil
			case schema.TimeoutDelete:
				return struct{}{}, "ready", nil
			}
		}
		if err != nil {
			return nil, "", err
		}
		if ready(p) {
			return p, "ready", nil
		}
		return p, "pending", nil
	})
}
func projectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteProject(ctx, d.Id()); err != nil {
		if cloapi.IsNotFound(err) {
			return nil
		}
		return diag.FromErr(err)
	}
	if err := waitProject(ctx, d, cli, schema.TimeoutDelete, func(p *cloapi.Project) bool { return p.Status == "DELETED" }); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
