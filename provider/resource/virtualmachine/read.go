package virtualmachine

import (
	"terraform-provider-vbridge/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Read(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(*api.Client)

	vmID := d.Id()
	vm, err := apiClient.GetVMDetailedByID(vmID)
	if err != nil {
		return err
	}

	d.Set("client_id", vm.ClientId)
	d.Set("name", vm.Name)
	d.Set("template", vm.Template)
	d.Set("guest_os_id", vm.GuestOsId)
	d.Set("cores", vm.Specification.Cores)
	d.Set("memory_size", vm.Specification.MemoryGb)
	d.Set("mo_ref", vm.Specification.MoRef)
	d.Set("operating_system_disk_capacity", int(vm.Specification.VirtualDisks[0].Capacity))
	d.Set("operating_system_disk_storage_profile", vm.Specification.VirtualDisks[0].Tier)
	if vm.MountedISO != nil {
		d.Set("iso_file", *vm.MountedISO)
	}
	d.Set("backup_type", vm.BackupType)
	d.Set("hosting_location_id", vm.Specification.HostingLocationId)
	d.Set("hosting_location_name", vm.HostingLocation.Name)
	d.Set("hosting_location_default_network", vm.HostingLocation.DefaultNetwork)
	d.Set("vm_id", vm.Id.String())

	var additionalDisks []interface{}
	for _, disk := range vm.Specification.VirtualDisks {
		additionalDisks = append(additionalDisks, map[string]interface{}{
			"capacity":        int(disk.Capacity),
			"storage_profile": disk.Tier,
		})
	}
	d.Set("additional_disks", additionalDisks)

	return nil
}
