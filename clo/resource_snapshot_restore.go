package clo

import (
	"context"

	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSnapshotRestore() *schema.Resource {
	return &schema.Resource{
		Description:   "Provision a server from a snapshot. Destroy deletes the server and applies the configured attachment deletion policy.",
		ReadContext:   resourceSnapshotRestoreRead,
		CreateContext: resourceSnapshotRestoreCreate,
		DeleteContext: resourceInstanceDelete,
		UpdateContext: resourceSnapshotRestoreRead,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"deletion": serverDeletionSchema(),
			"snapshot_id": {
				Description: "ID of the snapshot to restore from. The snapshot must be in the ACTIVE status.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Name of the server to provision from the snapshot",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"id": {
				Description: "ID of the provisioned server",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "Lifecycle status of the provisioned server",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project": {
				Description: "ID of the project the server belongs to",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"addresses": {
				Description: "Addresses attached to the provisioned server",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"created_in": {
				Description: "Timestamp the server was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceSnapshotRestoreCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := cli.RestoreSnapshot(ctx, d.Get("snapshot_id").(string), d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err := waitInstanceState(ctx, id, cli, []string{creatingInstance}, []string{activeInstance}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}
	return resourceSnapshotRestoreRead(ctx, d, m)
}

func resourceSnapshotRestoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	srv, err := cli.GetServer(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	// name is a write-only input (the server's name equals it) and is preserved
	// in state as configured.
	fields := map[string]interface{}{
		"id":         srv.ID,
		"status":     srv.Status,
		"project":    srv.Project,
		"addresses":  srv.Addresses,
		"created_in": srv.CreatedIn,
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}
