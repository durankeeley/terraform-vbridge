package virtualmachine

import (
	"terraform-provider-vbridge/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Create(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(*api.Client)
	vm := api.VirtualMachine{
		ClientId:   d.Get("client_id").(int),
		Name:       d.Get("name").(string),
		Template:   d.Get("template").(string),
		GuestOsId:  d.Get("guest_os_id").(string),
		Cores:      d.Get("cores").(int),
		MemorySize: d.Get("memory_size").(int),
		OperatingSystemDisk: api.Disk{
			Capacity:       d.Get("operating_system_disk_capacity").(int),
			StorageProfile: d.Get("operating_system_disk_storage_profile").(string),
		},
		BackupType: d.Get("backup_type").(string),
		HostingLocation: api.HostingLocation{
			Id:             d.Get("hosting_location_id").(string),
			Name:           d.Get("hosting_location_name").(string),
			DefaultNetwork: d.Get("hosting_location_default_network").(string),
		},
		QuoteItem: make(map[string]interface{}), // Initialize with an empty map
	}

	if v, ok := d.GetOk("iso_file"); ok {
		vm.IsoFile = v.(string)
	}

	if v, ok := d.GetOk("additional_disks"); ok {
		for _, disk := range v.([]interface{}) {
			d := disk.(map[string]interface{})
			vm.AdditionalDisks = append(vm.AdditionalDisks, api.Disk{
				Capacity:       d["capacity"].(int),
				StorageProfile: d["storage_profile"].(string),
			})
		}
	}

	if v, ok := d.GetOk("quote_item"); ok {
		vm.QuoteItem = v.(map[string]interface{})
	}

	vmID, err := apiClient.CreateVM(vm)
	if err != nil {
		return err
	}

	d.SetId(vmID)
	d.Set("vm_id", vmID)

	return Read(d, meta)
}
