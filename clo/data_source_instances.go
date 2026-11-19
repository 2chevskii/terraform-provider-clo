package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceInstances() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of the instances",
		ReadContext: dataSourceInstancesRead,
		Schema: map[string]*schema.Schema{
			"query":       listQuerySchema(),
			"total_count": computedField(schema.TypeInt),
			"project_id": {
				Description: "ID of the project that owns instances",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Type:        schema.TypeList,
				Description: "The object that holds the results",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the instance",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "Name of the instance",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"image": {
							Description: "ID of the image from which the instance was installed",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"recipe_id": {
							Description: "ID of the recipe that was installed on the instance",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"image_min_disk":   computedField(schema.TypeInt),
						"image_min_ram":    computedField(schema.TypeInt),
						"os_family":        computedField(schema.TypeString),
						"os_distribution":  computedField(schema.TypeString),
						"os_version":       computedField(schema.TypeString),
						"recipe_min_disk":  computedField(schema.TypeInt),
						"recipe_min_ram":   computedField(schema.TypeInt),
						"recipe_min_vcpus": computedField(schema.TypeInt),
						"snapshots":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
						"admin_key_status": computedField(schema.TypeString),
						"datacenter":       computedField(schema.TypeString),
						"primary_address":  computedField(schema.TypeString),
						"cpu_type":         computedField(schema.TypeString),
						"flavor_disk":      computedField(schema.TypeInt),
						"recipe_status":    computedField(schema.TypeString),
						"switch_status":    computedField(schema.TypeString),
						"project":          computedField(schema.TypeString),
						"guest_agent":      computedField(schema.TypeBool),
						"recipe_name":      computedField(schema.TypeString),
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_in": {
							Description: "Timestamp the instance was created",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"rescue_mode": {
							Description: "Describes is the instance in rescue mode",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"flavor_vcpus": {
							Description: "Number of VCPU of the instance",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"flavor_ram": {
							Description: "Amount of RAM of the instance",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"addresses": {
							Description: "Information about addresses attached to the instance",
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"disk_data": {
							Description: "Information about disks attached to the instance",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"id": {
									Description: "ID of the attached disk",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"storage_type": {
									Description: "Storage type of the attached disk. Could be `volume` or `local`",
									Type:        schema.TypeString,
									Computed:    true,
								},
							}}},
					},
				},
			},
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	options := listOptions(d)
	servers, err := cli.ListServers(ctx, d.Get("project_id").(string), options)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("total_count", options.Total); err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("result", flattenInstancesResults(servers)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenInstancesResults(pr []cloapi.Server) []interface{} {
	res := make([]interface{}, 0, len(pr))
	for _, p := range pr {
		res = append(res, map[string]interface{}{
			"id":               p.ID,
			"name":             p.Name,
			"image":            p.ImageName,
			"image_min_disk":   p.ImageMinDisk,
			"image_min_ram":    p.ImageMinRAM,
			"os_family":        p.OSFamily,
			"os_distribution":  p.OSDistribution,
			"os_version":       p.OSVersion,
			"recipe_min_disk":  p.RecipeMinDisk,
			"recipe_min_ram":   p.RecipeMinRAM,
			"recipe_min_vcpus": p.RecipeMinVCPUs,
			"snapshots":        p.Snapshots,
			"admin_key_status": p.AdminKeyStatus,
			"datacenter":       p.Datacenter,
			"primary_address":  p.PrimaryAddress,
			"cpu_type":         p.CPUType,
			"flavor_disk":      p.FlavorDisk,
			"recipe_status":    p.RecipeStatus,
			"switch_status":    p.SwitchStatus,
			"project":          p.Project,
			"guest_agent":      p.GuestAgent,
			"addresses":        p.Addresses,
			"recipe_name":      p.RecipeName,
			"status":           p.Status,
			"recipe_id":        p.RecipeName,
			"created_in":       p.CreatedIn,
			"rescue_mode":      p.RescueMode,
			"flavor_ram":       p.FlavorRam,
			"flavor_vcpus":     p.FlavorVcpus,
			"disk_data":        flattenServerDisks(p.Disks),
		})
	}
	return res
}
