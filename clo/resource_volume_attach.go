package clo

import (
	"context"
	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVolumeAttach() *schema.Resource {
	return &schema.Resource{
		Description:   "Attach the volume to an instance",
		ReadContext:   resourceVolumeAttachRead,
		CreateContext: resourceVolumeAttachCreate,
		DeleteContext: resourceVolumeDetach,
		UpdateContext: resourceVolumeAttachRead,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"volume_id": {
				Description: "ID of the volume which should be attached",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"instance_id": {
				Description: "ID of the instance to which the volume will be attached",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"force_detach": {Type: schema.TypeBool, Optional: true, Default: true, Description: "Force volume detach."},
			"clear_fstab":  {Type: schema.TypeBool, Optional: true, Description: "Clear the filesystem table entry on detach."},
			"device": {
				Description: "Represents the filesystem's device name, for example: `/dev/vdb`",
				Type:        schema.TypeString, Computed: true},
		},
	}
}

func resourceVolumeAttachCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vid := d.Get("volume_id").(string)
	sid := d.Get("instance_id").(string)
	cli := m.(*providerMeta).v3
	if err := cli.AttachVolume(ctx, vid, sid); err != nil {
		return diag.FromErr(err)
	}
	if err := waitVolumeState(ctx, vid, cli, []string{attachingVolume}, []string{attachedVolume}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}
	vol, err := cli.GetVolume(ctx, vid)
	if err != nil {
		return diag.FromErr(err)
	}
	if vol.Attachment != nil {
		if e := d.Set("device", vol.Attachment.Device); e != nil {
			return diag.FromErr(e)
		}
	}
	d.SetId(vid)
	return nil
}

func resourceVolumeAttachRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	vol, err := cli.GetVolume(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if vol.Attachment == nil || vol.Attachment.ID != d.Get("instance_id").(string) {
		d.SetId("")
		return nil
	}
	device := vol.Attachment.Device
	if e := d.Set("device", device); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceVolumeDetach(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vid := d.Id()
	cli := m.(*providerMeta).v3
	if err := cli.DetachVolume(ctx, vid, d.Get("force_detach").(bool), d.Get("clear_fstab").(bool)); err != nil {
		return diag.FromErr(err)
	}
	if err := waitVolumeState(ctx, vid, cli, []string{detachingVolume}, []string{activeVolume}, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
