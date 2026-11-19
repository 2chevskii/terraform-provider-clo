package clo

import (
	"context"
	"errors"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	activeInstance   = "ACTIVE"
	stoppedInstance  = "STOPPED"
	creatingInstance = "BUILDING"
	startingInstance = "STARTING"
	stoppingInstance = "STOPPING"
	resizingInstance = "RESIZING"
	deletingInstance = "DELETING"
	deletedInstance  = "DELETED"

	switchOnServer = "ON"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Description:   "Project compute instance",
		ReadContext:   resourceInstanceRead,
		CreateContext: resourceInstanceCreate,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"deletion": serverDeletionSchema(),
			"project_id": {
				Description: "ID of the project where the instance should be created",
				ForceNew:    true,
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name of the instance. Changing it renames the instance in place.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"password": {
				Description: "Password for the new instance",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"image_id": {
				Description: "Operating system image ID.", Type: schema.TypeString, Optional: true,
				ForceNew: true,
			},
			"volume_id": {Type: schema.TypeString, Optional: true, ForceNew: true, Description: "Existing boot volume ID."},
			"user_data": {Type: schema.TypeString, Optional: true, ForceNew: true, Sensitive: true, Description: "Base64-encoded cloud-init script."},
			"cpu_type":  {Type: schema.TypeString, Optional: true, Computed: true, ForceNew: true, Description: "CPU type of the instance."},
			"flavor_ram": {
				Description: "Amount of RAM of the new instance",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"flavor_vcpus": {
				Description: "Number of VCPU of the new instance",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"block_device": {
				Description: "Disk data for the new instance",
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bootable": {
							Description: "Is the disk bootable",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"storage_type": {
							Description: "Storage type of the new disk. Could be `volume` or `local`",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"size": {
							Description: "Requested size of the new disk",
							Type:        schema.TypeInt,
							Required:    true,
						},
					},
				},
			},
			"no_addresses": {Type: schema.TypeBool, Optional: true, ForceNew: true, Description: "Send an empty address list during server creation. By default CLO chooses the addresses when no address blocks are present.", ConflictsWith: []string{"addresses"}},
			"addresses": {ConflictsWith: []string{"no_addresses"},
				Description: "Addresses for the new instance",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external": {
							Description: "Should the new address be the external one",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"version": {
							Description: "Version of the new address. Could be `4` or `6`",
							Type:        schema.TypeInt,
							Optional:    true,
							ForceNew:    true,
						},
						"address_id": {
							Description: "Use an existing IP with a provided ID",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"ddos_protection": {
							Description: "Should the new address be protected from DDoS",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"bandwidth": {
							Description: "Max address bandwidth, must be 100 or 1024",
							Type:        schema.TypeInt,
							Optional:    true, ForceNew: true,
						},
					},
				},
			},
			"keypairs": {
				Description: "The list contains the SSH-keypairs IDs",
				Type:        schema.TypeList,
				Optional:    true, ForceNew: true,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"recipe_id": {
				Description: "ID of the recipe that will be installed on the instance",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"licenses": {
				Description: "The list contains licences that should be ordered with the instance",
				Type:        schema.TypeList,
				ForceNew:    true,
				Optional:    true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"addon": {
						Type:     schema.TypeString,
						Required: true,
					},
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"value": {
						Deprecated: "Set value on clo_compute_instance_license. The server creation endpoint accepts only addon and name.",
						ValidateFunc: func(interface{}, string) ([]string, []error) {
							return nil, []error{errors.New("use clo_compute_instance_license to set a license value")}
						},
						Type:     schema.TypeInt,
						Optional: true,
					},
				}},
			},
			"id": {
				Description: "ID of the created instance",
				Type:        schema.TypeString, Computed: true},
			"status": {
				Description: "Current status of the instance",
				Type:        schema.TypeString, Computed: true},
			"switch_status": {
				Description: "Power switch position reported by the API (`ON`/`OFF`)",
				Type:        schema.TypeString, Computed: true},
			"created_in": {
				Description: "Timestamp the instance was created",
				Type:        schema.TypeString, Computed: true},
		},
	}
}

// Actions

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := cli.CreateServer(ctx, buildServerCreateParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	if err := waitInstanceState(ctx, id, cli, []string{creatingInstance}, []string{activeInstance}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	if p, ok := d.GetOk("password"); ok {
		if err := cli.ChangeServerPassword(ctx, id, p.(string)); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	servID := d.Id()

	if d.HasChange("name") {
		if err := cli.RenameServer(ctx, servID, d.Get("name").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("flavor_ram", "flavor_vcpus") {
		_, ram := d.GetChange("flavor_ram")
		_, vcpus := d.GetChange("flavor_vcpus")
		if err := resizeServer(ctx, servID, vcpus.(int), ram.(int), cli, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("password") {
		_, c := d.GetChange("password")
		if err := cli.ChangeServerPassword(ctx, servID, c.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	srv, err := cli.GetServer(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	fields := map[string]interface{}{
		"id":            srv.ID,
		"name":          srv.Name,
		"status":        srv.Status,
		"switch_status": srv.SwitchStatus,
		"project_id":    srv.Project,
		"created_in":    srv.CreatedIn,
		"flavor_ram":    srv.FlavorRam, "flavor_vcpus": srv.FlavorVcpus, "cpu_type": srv.CPUType,
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

// Helpers

func buildServerCreateParams(d *schema.ResourceData) cloapi.ServerCreateParams {
	return cloapi.ServerCreateParams{
		ProjectID: d.Get("project_id").(string),
		Name:      d.Get("name").(string),
		ImageID:   optString(d, "image_id"),
		VolumeID:  optString(d, "volume_id"), UserData: optString(d, "user_data"), CPUType: optString(d, "cpu_type"),
		FlavorRam:   d.Get("flavor_ram").(int),
		FlavorVcpus: d.Get("flavor_vcpus").(int),
		RecipeID:    optString(d, "recipe_id"),
		Storages:    buildInstanceStorages(d),
		Addresses:   buildInstanceAddresses(d),
		Licenses:    buildInstanceLicenses(d),
		Keypairs:    buildInstanceKeypairs(d),
	}
}

func optString(d *schema.ResourceData, key string) string {
	if v, ok := d.GetOk(key); ok {
		return v.(string)
	}
	return ""
}

func buildInstanceKeypairs(d *schema.ResourceData) []string {
	ks, ok := d.GetOk("keypairs")
	if !ok {
		return nil
	}
	kp := ks.([]interface{})
	keyPairs := make([]string, len(kp))
	for i, v := range kp {
		keyPairs[i] = v.(string)
	}
	return keyPairs
}

func buildInstanceLicenses(d *schema.ResourceData) []cloapi.ServerLicense {
	v, ok := d.GetOk("licenses")
	if !ok {
		return nil
	}
	var out []cloapi.ServerLicense
	for _, l := range v.([]interface{}) {
		m := l.(map[string]interface{})
		lic := cloapi.ServerLicense{}
		if x, ok := m["addon"]; ok {
			lic.Addon = x.(string)
		}
		if x, ok := m["name"]; ok {
			lic.Name = x.(string)
		}
		out = append(out, lic)
	}
	return out
}

func buildInstanceStorages(d *schema.ResourceData) []cloapi.ServerStorage {
	v, ok := d.GetOk("block_device")
	if !ok {
		return nil
	}
	var out []cloapi.ServerStorage
	for i, stor := range v.([]any) {
		m := stor.(map[string]any)
		s := cloapi.ServerStorage{}
		if x, ok := m["size"]; ok {
			s.Size = x.(int)
		}
		if configuredBlockField(d, "block_device", i, "bootable") {
			value := m["bootable"].(bool)
			s.Bootable = &value
		}
		if x, ok := m["storage_type"]; ok {
			s.StorageType = x.(string)
		}
		out = append(out, s)
	}
	return out
}

func buildInstanceAddresses(d *schema.ResourceData) []cloapi.ServerAddress {
	entries := d.Get("addresses").([]interface{})
	if len(entries) == 0 && !d.Get("no_addresses").(bool) {
		return nil
	}
	out := make([]cloapi.ServerAddress, 0, len(entries))
	for i, adr := range entries {
		m := adr.(map[string]any)
		a := cloapi.ServerAddress{}
		if configuredBlockField(d, "addresses", i, "external") {
			value := m["external"].(bool)
			a.External = &value
		}
		if x, ok := m["address_id"]; ok {
			a.AddressID = x.(string)
		}
		if configuredBlockField(d, "addresses", i, "version") {
			value := m["version"].(int)
			a.Version = &value
		}
		if configuredBlockField(d, "addresses", i, "ddos_protection") {
			value := m["ddos_protection"].(bool)
			a.DdosProtection = &value
		}
		if x, ok := m["bandwidth"]; ok {
			a.Bandwidth = x.(int)
		}
		out = append(out, a)
	}
	return out
}

// SDKv2 fills omitted scalar fields with zero values. Inspect the configuration
// to distinguish an omitted boolean from an explicit false in the API request.
func configuredBlockField(d *schema.ResourceData, block string, index int, key string) bool {
	raw := d.GetRawConfig()
	if raw.IsNull() {
		return false
	}
	items := raw.GetAttr(block)
	if !items.IsKnown() {
		return true
	}
	if items.IsNull() {
		return false
	}
	elements := items.AsValueSlice()
	return index < len(elements) && !elements[index].GetAttr(key).IsNull()
}

// Waiters
func waitInstanceDeleted(ctx context.Context, serverId string, cli *cloapi.Client, timeout time.Duration) error {
	return waitForState(ctx, timeout, []string{deletingInstance}, []string{deletedInstance}, func() (interface{}, string, error) {
		srv, err := cli.GetServer(ctx, serverId)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedInstance, nil
		}
		if err != nil {
			return nil, "", err
		}
		return srv, srv.Status, nil
	})
}

func waitInstanceState(ctx context.Context, serverId string, cli *cloapi.Client, pending []string, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		srv, err := cli.GetServer(ctx, serverId)
		if err != nil {
			return nil, "", err
		}
		return srv, srv.Status, nil
	})
}

// waitInstanceEnabled waits for the instance to settle running (ACTIVE) or
// stopped (STOPPED) after a Start/Stop. Failure statuses are not in the pending
// set, so StateChangeConf surfaces them as a failure instead of hanging.
func waitInstanceEnabled(ctx context.Context, id string, cli *cloapi.Client, enabled bool, timeout time.Duration) error {
	if enabled {
		return waitInstanceState(ctx, id, cli, []string{stoppedInstance, startingInstance}, []string{activeInstance}, timeout)
	}
	return waitInstanceState(ctx, id, cli, []string{activeInstance, stoppingInstance}, []string{stoppedInstance}, timeout)
}

func resizeServer(ctx context.Context, id string, vcpus int, ram int, cli *cloapi.Client, d *schema.ResourceData) error {
	if err := cli.ResizeServer(ctx, id, ram, vcpus); err != nil {
		return err
	}
	return waitInstanceState(ctx, id, cli, []string{resizingInstance}, []string{activeInstance, stoppedInstance}, d.Timeout(schema.TimeoutUpdate))
}
