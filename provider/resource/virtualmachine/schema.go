package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"client_id": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"template": {
			Type:     schema.TypeString,
			Required: true,
		},
		"guest_os_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"cores": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"memory_size": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"operating_system_disk_capacity": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"operating_system_disk_storage_profile": {
			Type:     schema.TypeString,
			Required: true,
		},
		"additional_disks": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"capacity": {
						Type:     schema.TypeInt,
						Required: true,
					},
					"storage_profile": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
		"iso_file": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"quote_item": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"hosting_location_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"hosting_location_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"hosting_location_default_network": {
			Type:     schema.TypeString,
			Required: true,
		},
		"backup_type": {
			Type:     schema.TypeString,
			Required: true,
		},
		"vm_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"mo_ref": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}
