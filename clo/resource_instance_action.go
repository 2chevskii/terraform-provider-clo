package clo

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// An action receipt is retained across refresh. Only creation or replacement
// submits the request, since neither operation has a remote object to delete.
func resourceInstanceAction() *schema.Resource {
	return &schema.Resource{
		Description: "Submit a reboot or rescue request once. Change trigger to request another execution. Destroy forgets the receipt.",
		Schema: map[string]*schema.Schema{
			"instance_id": {Type: schema.TypeString, Required: true, ForceNew: true},
			"action":      {Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validation.StringInSlice([]string{"reboot", "rescue"}, false)},
			"trigger":     {Type: schema.TypeString, Optional: true, ForceNew: true, Description: "User-supplied revision for repeat executions."},
		},
		CreateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			cli := m.(*providerMeta).v3
			server := d.Get("instance_id").(string)
			action := d.Get("action").(string)
			var err error
			if action == "rescue" {
				err = cli.RescueServer(ctx, server)
			} else {
				err = cli.RebootServer(ctx, server)
			}
			if err != nil {
				return diag.FromErr(err)
			}
			d.SetId(server + "/" + action)
			return nil
		},
		ReadContext: schema.NoopContext, DeleteContext: schema.NoopContext,
	}
}
