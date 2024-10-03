package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type VirtualMachine struct {
	ClientId            int      `json:"clientId"`
	Name                string   `json:"name"`
	Template            string   `json:"template"`
	GuestOsId           string   `json:"guestOsId"`
	Cores               int      `json:"cores"`
	MemorySize          int      `json:"memorySize"`
	OperatingSystemDisk Disk     `json:"operatingSystemDisk"`
	AdditionalDisks     []Disk   `json:"additionalDisks,omitempty"`
	IsoFile             string   `json:"isoFile,omitempty"`
	QuoteItem           Quote    `json:"quoteItem,omitempty"`
	BackupType          string   `json:"backupType"`
	Id                  int      `json:"id"`
	HostingLocation     Location `json:"hostingLocation"`
	MoRef               string   `json:"moRef"`
}

type Disk struct {
	Capacity       int    `json:"capacity"`
	StorageProfile string `json:"storageProfile"`
}

type Quote struct{}

type Location struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	DefaultNetwork string `json:"defaultNetwork"`
}

var (
	vmMutex sync.Mutex
	vms     = make(map[int]VirtualMachine)
)

const (
	apiKey = "yourapikeygoeshere"
	user   = "you-users@yourcompany.com"
)

func main() {
	http.HandleFunc("/api/Provisioning/VirtualMachine", provisionVMHandler)
	http.HandleFunc("/api/client/virtualresources/", getAllVMsHandler)
	http.HandleFunc("/api/VirtualResource/Detailed/", getVMDetailedByIDHandler)

	log.Println("Starting server on :8087")
	log.Fatal(http.ListenAndServe(":8087", nil))
}

func authenticate(r *http.Request) bool {
	apiKeyHeader := r.Header.Get("Authorization")
	userHeader := r.Header.Get("x-mcs-user")

	return apiKeyHeader == fmt.Sprintf("Bearer %s", apiKey) && userHeader == user
}

func hasMissingRequiredFields(vm VirtualMachine) (bool, string) {
	fields := []struct {
		value string
		name  string
	}{
		{fmt.Sprint(vm.ClientId), "ClientId"},
		{vm.Name, "Name"},
		{vm.GuestOsId, "GuestOsId"},
		{fmt.Sprint(vm.Cores), "Cores"},
		{fmt.Sprint(vm.MemorySize), "MemorySize"},
		{vm.OperatingSystemDisk.StorageProfile, "OperatingSystemDisk.StorageProfile"},
		{vm.HostingLocation.Id, "HostingLocation.Id"},
		{vm.HostingLocation.Name, "HostingLocation.Name"},
		{vm.HostingLocation.DefaultNetwork, "HostingLocation.DefaultNetwork"},
		{vm.BackupType, "BackupType"},
	}

	for _, field := range fields {
		if field.value == "" || field.value == "0" {
			return true, field.name
		}
	}

	return false, ""
}

func provisionVMHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if !authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var vm VirtualMachine
	err := json.NewDecoder(r.Body).Decode(&vm)
	if err != nil {
		handleError(w, fmt.Sprintf("Error parsing VM file %s", err), http.StatusBadRequest)
		return
	}

	if missing, field := hasMissingRequiredFields(vm); missing {
		handleError(w, fmt.Sprintf("Missing required field: %s", field), http.StatusBadRequest)
		return
	}

	vmID := rand.Int()
	vm.Id = vmID
	vm.MoRef = uuid.New().String()

	if vm.Template == "Windows2022_Standard_30GB" {
		vm.OperatingSystemDisk.Capacity = 30
	}

	vmMutex.Lock()
	vms[vmID] = vm
	vmMutex.Unlock()

	if err := saveVMToFile(vmID, vm); err != nil {
		handleError(w, "Error saving VM details", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getAllVMsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if !authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	allVMs, err := loadAllVMs()
	if err != nil {
		handleError(w, "Error loading VMs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allVMs)
}

func getVMDetailedByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	vmID := path.Base(r.URL.Path)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		handleError(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	vm, err := loadVMFromFile(vmID)
	if err != nil {
		if os.IsNotExist(err) {
			handleError(w, "VM not found", http.StatusNotFound)
		} else {
			handleError(w, "Error reading VM file", http.StatusInternalServerError)
		}
		return
	}

	response := createVMDetailResponse(vm)
	log.Printf("Vm Detail Response %s", response)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		handleError(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}

func handleError(w http.ResponseWriter, message string, statusCode int) {
	log.Println(message)
	http.Error(w, message, statusCode)
}

func saveVMToFile(vmID int, vm VirtualMachine) error {
	file, err := json.MarshalIndent(vm, "", " ")
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%d.json", vmID)
	log.Printf("Saving VM file %d.json", vmID)
	return os.WriteFile(fileName, file, 0644)
}

func loadAllVMs() ([]map[string]interface{}, error) {
	var allVMs []map[string]interface{}
	log.Printf("Reading all json files")
	files, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			vm, err := loadVMFromFile(strings.TrimSuffix(file.Name(), ".json"))
			log.Printf("Reading VM file %s", file.Name())
			if err != nil {
				log.Printf("Error parsing VM file %s: %v\n", file.Name(), err)
				continue
			}

			vmMap := map[string]interface{}{
				"clientId":            vm.ClientId,
				"name":                vm.Name,
				"template":            vm.Template,
				"guestOsId":           vm.GuestOsId,
				"cores":               vm.Cores,
				"memorySize":          vm.MemorySize,
				"operatingSystemDisk": vm.OperatingSystemDisk,
				"additionalDisks":     vm.AdditionalDisks,
				"isoFile":             vm.IsoFile,
				"quoteItem":           vm.QuoteItem,
				"backupType":          vm.BackupType,
				"id":                  vm.Id,
				"hostingLocation":     vm.HostingLocation.Name,
			}

			allVMs = append(allVMs, vmMap)
		}
	}
	return allVMs, nil
}

func loadVMFromFile(vmID string) (VirtualMachine, error) {
	fileName := fmt.Sprintf("%s.json", vmID)
	data, err := os.ReadFile(fileName)
	log.Printf("Reading json file %s", fileName)
	if err != nil {
		return VirtualMachine{}, err
	}

	var vm VirtualMachine
	err = json.Unmarshal(data, &vm)
	if err != nil {
		return VirtualMachine{}, err
	}

	return vm, nil
}

func createVMDetailResponse(vm VirtualMachine) map[string]interface{} {
	log.Printf("Creating VM Detail Response")
	return map[string]interface{}{
		"clientId": vm.ClientId,
		"specification": map[string]interface{}{
			"cores":               vm.Cores,
			"sockets":             1,
			"memoryGb":            vm.MemorySize,
			"moRef":               vm.MoRef,
			"virtualDisks":        generateVirtualDisks(vm.OperatingSystemDisk, vm.AdditionalDisks),
			"backupType":          vm.BackupType,
			"hostingLocationName": vm.HostingLocation.Name,
			"hostingLocationId":   vm.HostingLocation.Id,
		},
		"id":                  vm.Id,
		"name":                vm.Name,
		"lastVirtualDisks":    1,
		"lastCPU":             vm.Cores,
		"lastMemory":          vm.MemorySize,
		"hostingLocation":     vm.HostingLocation.Name,
		"hostingLocationType": "DefaultType",
		"annotation":          "Default annotation",
	}
}

func generateVirtualDisks(operatingSystemDisk Disk, additionalDisks []Disk) []map[string]interface{} {
	var virtualDisks []map[string]interface{}
	log.Printf("Generating Virtual Disks for VM Detail Response")

	// Add the operating system disk as the first virtual disk
	virtualDisks = append(virtualDisks, map[string]interface{}{
		"moRef":              uuid.New().String(),
		"capacity":           operatingSystemDisk.Capacity,
		"vmfs":               operatingSystemDisk.StorageProfile,
		"slotInfo":           "slotInfo",
		"tier":               "tier",
		"name":               "OperatingSystemDisk",
		"capacityDesciption": "Operating system disk",
		"vDiskID":            "vDiskID",
		"filename":           "filename",
		"friendlyName":       "OS Disk",
	})

	// Add additional disks
	for _, disk := range additionalDisks {
		virtualDisk := map[string]interface{}{
			"moRef":              uuid.New().String(),
			"capacity":           disk.Capacity,
			"vmfs":               disk.StorageProfile,
			"slotInfo":           "slotInfo",
			"tier":               "tier",
			"name":               "AdditionalDisk",
			"capacityDesciption": "Additional disk",
			"vDiskID":            "vDiskID",
			"filename":           "filename",
			"friendlyName":       "Additional Disk",
		}
		virtualDisks = append(virtualDisks, virtualDisk)
		log.Printf("Added virtual disk: %+v", virtualDisk)
	}

	log.Printf("Generated Virtual Disks: %v", virtualDisks)
	return virtualDisks
}
