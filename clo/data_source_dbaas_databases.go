package clo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbaasDatabases() *schema.Resource {
	return &schema.Resource{
		Description: "Lists DBaaS databases in a project or cluster.",
		ReadContext: dataSourceDbaasDatabasesRead,
		Schema: map[string]*schema.Schema{
			"query":       listQuerySchema(),
			"total_count": computedField(schema.TypeInt),
			"cluster_id": {
				Description: "ID of the dbaas cluster that owns the databases",
				Type:        schema.TypeString,
				Optional:    true, ExactlyOneOf: []string{"cluster_id", "project_id"},
			},
			"project_id": {Type: schema.TypeString, Optional: true, ExactlyOneOf: []string{"cluster_id", "project_id"}, Description: "ID of the project containing the databases."},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeString, Computed: true},
						"name":           {Type: schema.TypeString, Computed: true},
						"cluster_id":     {Type: schema.TypeString, Computed: true},
						"project":        {Type: schema.TypeString, Computed: true},
						"admin_username": {Type: schema.TypeString, Computed: true},
						"status":         {Type: schema.TypeString, Computed: true},
						"backup_enabled": {Type: schema.TypeBool, Computed: true},
						"created_in":     {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceDbaasDatabasesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	scope := d.Get("cluster_id").(string)
	list := cli.ListDatabasesByCluster
	if project, ok := d.GetOk("project_id"); ok {
		scope = project.(string)
		list = cli.ListDatabasesByProject
	}
	options := listOptions(d)
	databases, err := list(ctx, scope, options)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("total_count", options.Total); err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(databases))
	for _, db := range databases {
		res = append(res, map[string]interface{}{
			"id":             db.ID,
			"name":           db.Name,
			"cluster_id":     db.ClusterID,
			"project":        db.Project,
			"admin_username": db.AdminUsername,
			"status":         db.Status,
			"backup_enabled": db.BackupEnabled,
			"created_in":     db.CreatedIn,
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(scope)
	return nil
}
