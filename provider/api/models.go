package api

import "encoding/json"

type VirtualMachine struct {
	ClientId            int                    `json:"clientId"`
	Name                string                 `json:"name"`
	Template            string                 `json:"template"`
	GuestOsId           string                 `json:"guestOsId"`
	Cores               int                    `json:"cores"`
	MemorySize          int                    `json:"memorySize"`
	OperatingSystemDisk Disk                   `json:"operatingSystemDisk"`
	AdditionalDisks     []Disk                 `json:"additionalDisks,omitempty"`
	IsoFile             string                 `json:"isoFile,omitempty"`
	QuoteItem           map[string]interface{} `json:"quoteItem"`
	HostingLocation     HostingLocation        `json:"hostingLocation"`
	BackupType          string                 `json:"backupType"`
	Id                  json.Number            `json:"id,omitempty"`
	Specification       Specification          `json:"specification"`
	MountedISO          *string                `json:"mountedISO"`
}

type HostingLocation struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	DefaultNetwork string `json:"defaultNetwork"`
}

type Specification struct {
	HealthState       string          `json:"healthState"`
	PowerState        string          `json:"powerState"`
	Cores             int             `json:"cores"`
	Sockets           int             `json:"sockets"`
	MemoryGb          int             `json:"memoryGb"`
	MoRef             string          `json:"moRef"`
	VirtualDisks      []VirtualDisk   `json:"virtualDisks"`
	NetworkDevices    []NetworkDevice `json:"networkDevices"`
	HostingLocationId string          `json:"hostingLocationId"`
}

type VirtualDisk struct {
	MoRef               string  `json:"moRef"`
	Capacity            float64 `json:"capacity"`
	Vmfs                string  `json:"vmfs"`
	SlotInfo            string  `json:"slotInfo"`
	Tier                string  `json:"tier"`
	Name                string  `json:"name"`
	CapacityDescription string  `json:"capacityDescription"`
}

type NetworkDevice struct {
	Name           string `json:"name"`
	MoRef          string `json:"moRef"`
	NetworkName    string `json:"networkName"`
	MacAddress     string `json:"macAddress"`
	Connected      bool   `json:"connected"`
	StartConnected bool   `json:"startConnected"`
	NetworkId      string `json:"networkId"`
}

type Disk struct {
	Capacity       int    `json:"capacity"`
	StorageProfile string `json:"storageProfile"`
}
