package api

type VirtualMachineLookup interface {
    GetVMByName(vmName string, clientId int) (string, error)
}

type DiskManager interface {
    CreateAdditionalDisk(vmID string, disk VirtualDisk) error
    GetAdditionalDisk(vmID string, diskID string) (*VirtualDisk, error)
}
