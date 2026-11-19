package clo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	creatingVolume  = "CREATING"
	activeVolume    = "AVAILABLE"
	resizingVolume  = "RESIZING"
	attachingVolume = "ATTACHING"
	attachedVolume  = "IN_USE"
	detachingVolume = "DETACHING"
	deletingVolume  = "DELETING"
	deletedVolume   = "DELETED"
)

func resourceVolume() *schema.Resource {
	return &schema.Resource{
		Importer:      scopedImporter("project_id"),
		Description:   "Create a new volume in the project",
		ReadContext:   resourceVolumeRead,
		CreateContext: resourceVolumeCreate,
		UpdateContext: resourceVolumeUpdate,
		DeleteContext: resourceVolumeDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("size", func(ctx context.Context, oldValue, newValue, meta interface{}) error {
				if newValue.(int) < oldValue.(int) {
					return errors.New("size could be increased only")
				}
				return nil
			})),
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project where the volume should be created",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human-readable name of the new volume",
				Type:        schema.TypeString,
				Optional:    true, Computed: true,
			},
			"description":   {Type: schema.TypeString, Optional: true, Description: "Volume description."},
			"autorename":    {Type: schema.TypeBool, Optional: true, ForceNew: true, Description: "Request automatic naming at create. Omitting name also enables automatic naming."},
			"force_delete":  {Type: schema.TypeBool, Optional: true, Description: "Force volume deletion."},
			"clear_fstab":   {Type: schema.TypeBool, Optional: true, Description: "Clear the filesystem table entry on deletion."},
			"reboot_server": {Type: schema.TypeBool, Optional: true, Description: "Request a server reboot when extending an attached volume."},
			"size": {
				Description: "Size of the new volume in Gb",
				Type:        schema.TypeInt,
				Required:    true,
				ValidateFunc: func(i interface{}, s string) (warns []string, errs []error) {
					if sz := i.(int); sz < 10 {
						errs = append(errs, errors.New("size should be at least 10Gb"))
					}
					return
				},
			},
			"id": {
				Description: "ID of the new volume",
				Type:        schema.TypeString, Computed: true},
			"status": {Type: schema.TypeString, Computed: true},
			"created_in": {
				Description: "Timestamp the volume was created",
				Type:        schema.TypeString, Computed: true},
		},
	}
}

func resourceVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := createVolume(ctx, cli, d)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	if err := waitVolumeState(ctx, id, cli, []string{creatingVolume}, []string{activeVolume}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}
	if desc, ok := d.GetOk("description"); ok {
		value := desc.(string)
		if err := cli.UpdateVolume(ctx, id, nil, &value); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	var name, description *string
	if d.HasChange("name") {
		v := d.Get("name").(string)
		name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		description = &v
	}
	if name != nil || description != nil {
		if err := cli.UpdateVolume(ctx, d.Id(), name, description); err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("size") {
		_, c := d.GetChange("size")
		if err := resizeVolume(ctx, cli, d, c.(int)); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteVolume(ctx, d.Id(), d.Get("force_delete").(bool), d.Get("clear_fstab").(bool)); err != nil {
		return diag.FromErr(err)
	}
	if err := waitVolumeDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	vol, err := cli.GetVolume(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	for k, v := range map[string]interface{}{"name": vol.Name, "description": vol.Description, "size": vol.Size} {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}
	if e := d.Set("id", vol.ID); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", vol.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("created_in", vol.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

// Waiters
func waitVolumeState(ctx context.Context, id string, cli *cloapi.Client, pending []string, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		vol, err := cli.GetVolume(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return vol, strings.ToUpper(vol.Status), nil
	})
}

func waitVolumeDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	return waitForState(ctx, timeout, []string{deletingVolume}, []string{deletedVolume}, func() (interface{}, string, error) {
		vol, err := cli.GetVolume(ctx, id)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedVolume, nil
		}
		if err != nil {
			return nil, "", err
		}
		return vol, strings.ToUpper(vol.Status), nil
	})
}

// Api actions
func createVolume(ctx context.Context, cli *cloapi.Client, d *schema.ResourceData) (string, error) {
	p := cloapi.VolumeCreateParams{
		ProjectID: d.Get("project_id").(string),
		Size:      d.Get("size").(int), Autorename: d.Get("autorename").(bool),
	}
	if n, ok := d.GetOk("name"); ok {
		p.Name = n.(string)
	} else {
		p.Autorename = true
	}
	return cli.CreateVolume(ctx, p)
}

func resizeVolume(ctx context.Context, cli *cloapi.Client, d *schema.ResourceData, size int) error {
	if err := cli.ExtendVolume(ctx, d.Id(), size, d.Get("reboot_server").(bool)); err != nil {
		return err
	}
	return waitVolumeState(ctx, d.Id(), cli, []string{resizingVolume}, []string{activeVolume, attachedVolume}, d.Timeout(schema.TimeoutUpdate))
}
