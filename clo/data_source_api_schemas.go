package clo

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceAccountBalance() *schema.Resource {
	return apiDataSource("AccountBalance", "", "Account balance (/v2/balance).", false, true, recordFields(map[string]*schema.Schema{
		"balance":      computedField(schema.TypeFloat),
		"credit_limit": computedField(schema.TypeFloat),
		"cost": recordFields(map[string]*schema.Schema{
			"hour":  computedField(schema.TypeFloat),
			"month": computedField(schema.TypeFloat),
		}),
		"cost_active": recordFields(map[string]*schema.Schema{
			"hour":  computedField(schema.TypeFloat),
			"month": computedField(schema.TypeFloat),
		}),
		"days_left_until_zero": computedField(schema.TypeInt),
	}))
}

func dataSourceAccountLimits() *schema.Resource {
	return apiDataSource("AccountLimits", "", "Account limits list (/v2/limits).", true, true, recordFields(map[string]*schema.Schema{
		"id":            computedField(schema.TypeInt),
		"name":          computedField(schema.TypeString),
		"measure":       computedField(schema.TypeString),
		"used":          computedField(schema.TypeInt),
		"max":           computedField(schema.TypeInt),
		"type":          computedField(schema.TypeString),
		"target_entity": computedField(schema.TypeString),
	}))
}

func dataSourceAccountProjectsLimits() *schema.Resource {
	return apiDataSource("AccountProjectsLimits", "", "Get limits for all projects (/v2/limits/projects).", true, true, recordFields(map[string]*schema.Schema{
		"id":            computedField(schema.TypeInt),
		"name":          computedField(schema.TypeString),
		"measure":       computedField(schema.TypeString),
		"used":          computedField(schema.TypeInt),
		"max":           computedField(schema.TypeInt),
		"type":          computedField(schema.TypeString),
		"target_entity": computedField(schema.TypeString),
		"project_id":    computedField(schema.TypeString),
		"enabled":       computedField(schema.TypeBool),
	}))
}

func dataSourceAccountStat() *schema.Resource {
	return apiDataSource("AccountStat", "", "Account stat (/v2/stat).", true, true, recordFields(map[string]*schema.Schema{
		"project_id": computedField(schema.TypeString),
		"param":      computedField(schema.TypeString),
		"addon":      computedField(schema.TypeString),
		"statdate":   computedField(schema.TypeString),
		"value":      computedField(schema.TypeFloat),
		"measure":    computedField(schema.TypeString),
	}))
}

func dataSourceProjectDetail() *schema.Resource {
	return apiDataSource("ProjectDetail", "project_id", "Project detail (/v2/projects/{object_id}/detail).", false, true, recordFields(map[string]*schema.Schema{
		"id":              computedField(schema.TypeString),
		"name":            computedField(schema.TypeString),
		"created":         computedField(schema.TypeString),
		"status":          computedField(schema.TypeString),
		"stopping_reason": computedField(schema.TypeString),
		"has_abuse":       computedField(schema.TypeBool),
		"switch_status":   computedField(schema.TypeString),
		"display_name":    computedField(schema.TypeString),
		"description":     computedField(schema.TypeString),
	}))
}

func dataSourceProjectLimitsList() *schema.Resource {
	return apiDataSource("ProjectLimitsList", "project_id", "Project limits list (/v2/projects/{object_id}/limits).", true, true, recordFields(map[string]*schema.Schema{
		"name":    computedField(schema.TypeString),
		"measure": computedField(schema.TypeString),
		"max":     computedField(schema.TypeInt),
		"used":    computedField(schema.TypeInt),
		"enabled": computedField(schema.TypeBool),
	}))
}

func dataSourceProjectServerConfigDep() *schema.Resource {
	return apiDataSource("ProjectServerConfigDep", "project_id", "Project server config dependencies (/v2/projects/{object_id}/servers/related-resources).", true, true, recordFields(map[string]*schema.Schema{
		"target":            computedField(schema.TypeString),
		"dependent":         computedField(schema.TypeString),
		"target_measure":    computedField(schema.TypeString),
		"dependent_measure": computedField(schema.TypeString),
		"min_target_value":  computedField(schema.TypeInt),
		"max_target_value":  computedField(schema.TypeInt),
		"dependencies": recordFields(map[string]*schema.Schema{
			"value":         computedField(schema.TypeInt),
			"min_dependent": computedField(schema.TypeInt),
		}),
	}))
}

func dataSourceProjectDbaasConfigDep() *schema.Resource {
	return apiDataSource("ProjectDbaasConfigDep", "project_id", "Project Dbaas config dependencies (/v2/projects/{object_id}/dbaas/related-resources).", true, true, recordFields(map[string]*schema.Schema{
		"target":            computedField(schema.TypeString),
		"dependent":         computedField(schema.TypeString),
		"target_measure":    computedField(schema.TypeString),
		"dependent_measure": computedField(schema.TypeString),
		"min_target_value":  computedField(schema.TypeInt),
		"max_target_value":  computedField(schema.TypeInt),
		"dependencies": recordFields(map[string]*schema.Schema{
			"value":         computedField(schema.TypeInt),
			"min_dependent": computedField(schema.TypeInt),
		}),
	}))
}

func dataSourceProjectInfrastructureModuleConstants() *schema.Resource {
	return apiDataSource("ProjectInfrastructureModuleConstants", "project_id", "Project infrastructure module constants (/v2/projects/{object_id}/params).", false, false, recordFields(map[string]*schema.Schema{
		"days_number_to_keep_dbaas_cluster_backups":  computedField(schema.TypeInt),
		"days_number_to_keep_dbaas_database_backups": computedField(schema.TypeInt),
		"max_addresses_per_server":                   computedField(schema.TypeInt),
		"max_snapshots_per_server":                   computedField(schema.TypeInt),
		"days_before_remove_ip":                      computedField(schema.TypeInt),
	}))
}

func dataSourceProjectConsumption() *schema.Resource {
	return apiDataSource("ProjectConsumption", "project_id", "Project consumption (/v2/projects/{object_id}/consumption).", true, true, recordFields(map[string]*schema.Schema{
		"name":          computedField(schema.TypeString),
		"amount":        computedField(schema.TypeFloat),
		"amount_active": computedField(schema.TypeFloat),
		"cost": recordFields(map[string]*schema.Schema{
			"hour":  computedField(schema.TypeFloat),
			"month": computedField(schema.TypeFloat),
		}),
		"cost_origin": recordFields(map[string]*schema.Schema{
			"hour":  computedField(schema.TypeFloat),
			"month": computedField(schema.TypeFloat),
		}),
		"discount_type":      computedField(schema.TypeString),
		"discount":           computedField(schema.TypeFloat),
		"discount_origin_by": computedField(schema.TypeString),
	}))
}

func dataSourceNetworksList() *schema.Resource {
	return apiDataSource("NetworksList", "project_id", "Networks list for project (/v2/projects/{object_id}/networks).", true, true, recordFields(map[string]*schema.Schema{
		"id":              computedField(schema.TypeString),
		"external":        computedField(schema.TypeBool),
		"is_private":      computedField(schema.TypeBool),
		"ddos_protection": computedField(schema.TypeBool),
		"subnet_address":  computedField(schema.TypeString),
		"gateway":         computedField(schema.TypeString),
		"status":          computedField(schema.TypeString),
		"mode":            computedField(schema.TypeString),
	}))
}

func dataSourceProjectLocalDisksList() *schema.Resource {
	return apiDataSource("ProjectLocalDisksList", "project_id", "Project local disks list (/v2/projects/{object_id}/local-disks).", true, true, recordFields(map[string]*schema.Schema{
		"id":          computedField(schema.TypeString),
		"name":        computedField(schema.TypeString),
		"created_in":  computedField(schema.TypeString),
		"status":      computedField(schema.TypeString),
		"size":        computedField(schema.TypeInt),
		"file_system": computedField(schema.TypeString),
		"bootable":    computedField(schema.TypeBool),
		"attached_to_server": recordFields(map[string]*schema.Schema{
			"id":     computedField(schema.TypeString),
			"device": computedField(schema.TypeString),
		}),
	}))
}

func dataSourceLocalDiskDetail() *schema.Resource {
	return apiDataSource("LocalDiskDetail", "disk_id", "Local disk detail (/v2/local-disks/{object_id}/detail).", false, true, recordFields(map[string]*schema.Schema{
		"id":          computedField(schema.TypeString),
		"name":        computedField(schema.TypeString),
		"created_in":  computedField(schema.TypeString),
		"status":      computedField(schema.TypeString),
		"size":        computedField(schema.TypeInt),
		"file_system": computedField(schema.TypeString),
		"bootable":    computedField(schema.TypeBool),
		"attached_to_server": recordFields(map[string]*schema.Schema{
			"id":     computedField(schema.TypeString),
			"device": computedField(schema.TypeString),
		}),
	}))
}

func dataSourceAvailableLicensesList() *schema.Resource {
	return apiDataSource("AvailableLicensesList", "", "Available licenses list (/v2/licenses).", true, true, recordFields(map[string]*schema.Schema{
		"addon": computedField(schema.TypeString),
		"items": recordFields(map[string]*schema.Schema{
			"name":  computedField(schema.TypeString),
			"title": computedField(schema.TypeString),
		}),
		"fields_json": computedField(schema.TypeString),
	}))
}

func dataSourceServerLicenses() *schema.Resource {
	return apiDataSource("ServerLicenses", "instance_id", "Server licenses list (/v2/servers/{object_id}/licenses).", true, true, recordFields(map[string]*schema.Schema{
		"id":           computedField(schema.TypeString),
		"addon":        computedField(schema.TypeString),
		"name":         computedField(schema.TypeString),
		"status":       computedField(schema.TypeString),
		"label":        computedField(schema.TypeString),
		"activated_in": computedField(schema.TypeString),
		"created_in":   computedField(schema.TypeString),
		"value":        computedField(schema.TypeInt),
	}))
}

func dataSourceLicenseDetails() *schema.Resource {
	return apiDataSource("LicenseDetails", "license_id", "License details (/v2/licenses/{object_id}).", false, true, recordFields(map[string]*schema.Schema{
		"id":           computedField(schema.TypeString),
		"addon":        computedField(schema.TypeString),
		"name":         computedField(schema.TypeString),
		"status":       computedField(schema.TypeString),
		"label":        computedField(schema.TypeString),
		"activated_in": computedField(schema.TypeString),
		"created_in":   computedField(schema.TypeString),
		"value":        computedField(schema.TypeInt),
	}))
}

func dataSourceProjectRuleList() *schema.Resource {
	return apiDataSource("ProjectRuleList", "project_id", "Project rules list (/v2/projects/{object_id}/loadbalancers/rules).", true, true, recordFields(map[string]*schema.Schema{
		"id":                     computedField(schema.TypeString),
		"external_protocol_port": computedField(schema.TypeInt),
		"internal_protocol_port": computedField(schema.TypeInt),
		"address":                computedField(schema.TypeString),
		"loadbalancer":           computedField(schema.TypeString),
		"server":                 computedField(schema.TypeString),
		"status":                 computedField(schema.TypeString),
	}))
}

func dataSourceLoadBalancerStat() *schema.Resource {
	return apiDataSource("LoadBalancerStat", "loadbalancer_id", "Get loadbalancer stat (/v2/loadbalancers/{object_id}/stat).", false, true, recordFields(map[string]*schema.Schema{
		"bytes_in":         computedField(schema.TypeInt),
		"bytes_out":        computedField(schema.TypeInt),
		"current_sessions": computedField(schema.TypeInt),
		"max_sessions":     computedField(schema.TypeInt),
	}))
}

func dataSourceMaintenanceModeStatus() *schema.Resource {
	return apiDataSource("MaintenanceModeStatus", "", "Get maintenance mode (/v2/maintenance-mode).", false, false, recordFields(map[string]*schema.Schema{
		"enable":           computedField(schema.TypeBool),
		"reason":           computedField(schema.TypeString),
		"message_template": computedField(schema.TypeString),
		"start_time":       computedField(schema.TypeString),
		"end_time":         computedField(schema.TypeString),
		"feedback_form":    computedField(schema.TypeBool),
		"message":          computedField(schema.TypeString),
	}))
}

func dataSourceInstanceGuestAgent() *schema.Resource {
	return apiDataSource("ServerCheckGuestAgent", "instance_id", "Check whether the instance guest agent is enabled.", false, false, recordFields(map[string]*schema.Schema{"guest_agent_enabled": computedField(schema.TypeBool)}))
}
func dataSourceInstanceConsole() *schema.Resource {
	result := recordFields(map[string]*schema.Schema{"console": recordFields(map[string]*schema.Schema{"type": computedField(schema.TypeString), "url": computedField(schema.TypeString)})})
	result.Sensitive = true
	return apiDataSource("ServerConsole", "instance_id", "Request a console session for this instance. The console URL is sensitive and is stored in Terraform state.", false, false, result)
}
