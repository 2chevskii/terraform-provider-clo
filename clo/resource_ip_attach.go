package clo

import (
	"context"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceIpAttach() *schema.Resource {
	return &schema.Resource{
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Description:   "Attach an address to the entity, for example: a loadbalancer or a server",
		ReadContext:   resourceIpAttachRead,
		CreateContext: resourceIpAttachCreate,
		UpdateContext: resourceIpAttachUpdate,
		DeleteContext: resourceIpDetach,
		Timeouts: &schema.ResourceTimeout{
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"address_id": {
				Description: "ID of the attached address",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"entity_id": {
				Description: "ID of the entity the address will be attached to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"entity_name": {
				Description: "Name of the entity. Should be `loadbalancer` or `server`",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"bandwidth": addressBandwidthSchema(),
			"is_primary": {
				Description: "Use the address as a primary address",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"status":  {Type: schema.TypeString, Computed: true},
			"address": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceIpAttachCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	adId := d.Get("address_id").(string)
	cli := m.(*providerMeta).v3

	if err := cli.AttachAddress(ctx, adId, d.Get("entity_id").(string), d.Get("entity_name").(string), d.Get("bandwidth").(int)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(adId)
	if err := waitAddressState(ctx, adId, cli, []string{processingIp}, []string{attachedIp}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("is_primary"); ok {
		if err := makePrimary(ctx, adId, cli, d); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIpAttachRead(ctx, d, m)
}

func resourceIpAttachRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	addr, err := cli.GetAddress(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	target := d.Get("entity_id").(string)
	kind := d.Get("entity_name").(string)
	if addr.AttachedTo == nil || (target != "" && addr.AttachedTo.ID != target) || (kind != "" && addr.AttachedTo.Entity != kind) {
		d.SetId("")
		return nil
	}
	for k, v := range map[string]string{"entity_id": addr.AttachedTo.ID, "entity_name": addr.AttachedTo.Entity, "address_id": addr.ID} {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("bandwidth", addr.Bandwidth); err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("status", addr.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("address", addr.Address); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("is_primary", addr.IsPrimary); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceIpDetach(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DetachAddress(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitAddressState(ctx, d.Id(), cli, []string{processingIp}, []string{detachedIp}, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceIpAttachUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("is_primary") && !d.Get("is_primary").(bool) {
		return diag.Errorf("CLO demotes a primary address when another address becomes primary. Set is_primary on the replacement address first.")
	}
	if d.HasChange("bandwidth") {
		if err := updateAddressBandwidth(ctx, d, m.(*providerMeta).v3, d.Get("bandwidth").(int)); err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("is_primary") {
		if err := makePrimary(ctx, d.Id(), m.(*providerMeta).v3, d); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceIpAttachRead(ctx, d, m)
}

func makePrimary(ctx context.Context, adId string, cli *cloapi.Client, d *schema.ResourceData) error {
	if err := cli.SetAddressPrimary(ctx, adId); err != nil {
		return err
	}
	return waitAddressState(ctx, adId, cli, []string{processingIp}, []string{attachedIp}, d.Timeout(schema.TimeoutUpdate))
}
