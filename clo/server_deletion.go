package clo

import (
	"context"
	"strings"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func serverDeletionSchema() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList, Optional: true, MaxItems: 1,
		Description: "Controls attachment cleanup on destroy. Omitting this block deletes all currently attached addresses and network volumes. An empty block retains them all.",
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"volume_ids":  {Type: schema.TypeSet, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Attached network volume IDs to delete."},
			"address_ids": {Type: schema.TypeSet, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Attached address IDs to delete."},
			"clear_fstab": {Type: schema.TypeBool, Optional: true, Description: "Clear filesystem table entries during deletion."},
		}},
	}
}

func serverDeletion(d *schema.ResourceData, srv *cloapi.Server) (addresses, volumes []string, clearFstab bool) {
	for _, disk := range srv.Disks {
		if strings.EqualFold(disk.StorageType, "volume") {
			volumes = append(volumes, disk.ID)
		}
	}
	blocks := d.Get("deletion").([]interface{})
	if len(blocks) == 0 {
		return srv.Addresses, volumes, false
	}
	if blocks[0] == nil {
		return []string{}, []string{}, false
	}
	policy := blocks[0].(map[string]interface{})
	selectIDs := func(attached []string, key string) []string {
		selected := []string{}
		ids := policy[key].(*schema.Set)
		for _, id := range attached {
			if ids.Contains(id) {
				selected = append(selected, id)
			}
		}
		return selected
	}
	return selectIDs(srv.Addresses, "address_ids"), selectIDs(volumes, "volume_ids"), policy["clear_fstab"].(bool)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	srv, err := cli.GetServer(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	addresses, volumes, clearFstab := serverDeletion(d, srv)
	if err := cli.DeleteServer(ctx, d.Id(), addresses, volumes, clearFstab); err != nil && !cloapi.IsNotFound(err) {
		return diag.FromErr(err)
	}
	if err := waitInstanceDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
