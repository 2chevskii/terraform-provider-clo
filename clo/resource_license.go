package clo

import (
	"context"
	"fmt"
	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
	"time"
)

func resourceInstanceLicense() *schema.Resource {
	return &schema.Resource{
		Description: "Manage a license installed on a server. Import using instance_id/license_id.",
		Schema: map[string]*schema.Schema{
			"instance_id": {Type: schema.TypeString, Required: true, ForceNew: true},
			"addon":       {Type: schema.TypeString, Required: true, ForceNew: true},
			"name":        {Type: schema.TypeString, Required: true},
			"value":       {Type: schema.TypeInt, Optional: true, Computed: true},
			"label":       computedField(schema.TypeString), "status": computedField(schema.TypeString),
			"created_in": computedField(schema.TypeString), "activated_in": computedField(schema.TypeString),
		},
		Timeouts: &schema.ResourceTimeout{Default: schema.DefaultTimeout(20 * time.Minute)},
		Importer: &schema.ResourceImporter{StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
			server, id, ok := strings.Cut(d.Id(), "/")
			if !ok || server == "" || id == "" {
				return nil, fmt.Errorf("expected instance_id/license_id")
			}
			if err := d.Set("instance_id", server); err != nil {
				return nil, err
			}
			d.SetId(id)
			return []*schema.ResourceData{d}, nil
		}},
		CreateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			cli := m.(*providerMeta).v3
			value := configuredInt(d, "value")
			id, err := cli.AddServerLicense(ctx, d.Get("instance_id").(string), d.Get("addon").(string), d.Get("name").(string), value)
			if err != nil {
				return diag.FromErr(err)
			}
			d.SetId(id)
			return awaitLicense(ctx, d, m, value, schema.TimeoutCreate)
		},
		ReadContext: licenseRead,
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			value := configuredInt(d, "value")
			if err := m.(*providerMeta).v3.UpdateLicense(ctx, d.Id(), d.Get("name").(string), value); err != nil {
				return diag.FromErr(err)
			}
			return awaitLicense(ctx, d, m, value, schema.TimeoutUpdate)
		},
		DeleteContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			err := m.(*providerMeta).v3.DeleteLicense(ctx, d.Id())
			if err != nil && !cloapi.IsNotFound(err) {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}
func licenseFields(ctx context.Context, d *schema.ResourceData, m interface{}) (map[string]interface{}, error) {
	rows, _, err := readQuery(ctx, m.(*providerMeta).v3, "LicenseDetails", d.Id(), nil, false, true)
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}
func licenseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fields, err := licenseFields(ctx, d, m)
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	for _, key := range []string{"addon", "name", "value", "label", "status", "created_in", "activated_in"} {
		if err := d.Set(key, fields[key]); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
func awaitLicense(ctx context.Context, d *schema.ResourceData, m interface{}, value *int, timeout string) diag.Diagnostics {
	err := waitForState(ctx, d.Timeout(timeout), []string{"pending"}, []string{"ready"}, func() (interface{}, string, error) {
		row, err := licenseFields(ctx, d, m)
		if timeout == schema.TimeoutCreate && cloapi.IsNotFound(err) {
			return struct{}{}, "pending", nil
		}
		if err != nil {
			return nil, "", err
		}
		if strings.Contains(fmt.Sprint(row["status"]), "ERROR") {
			return nil, "", fmt.Errorf("license %s entered %v", d.Id(), row["status"])
		}
		if row["name"] == d.Get("name") && (value == nil || fmt.Sprint(row["value"]) == fmt.Sprint(*value)) {
			return row, "ready", nil
		}
		return row, "pending", nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return licenseRead(ctx, d, m)
}
