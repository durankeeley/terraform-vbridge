package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	HostingLocation     Location `json:"hostingLocation"`
	BackupType          string   `json:"backupType"`
	Id                  string   `json:"id"`
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
	vms     = make(map[string]VirtualMachine)
)

const (
	apiKey = "yourapikeygoeshere"
	user   = "you-users@yourcompany.com"
)

func main() {
	http.HandleFunc("/api/Provisioning/VirtualMachine", provisionVMHandler)
	http.HandleFunc("/api/client/virtualresources/", getAllVMsHandler)

	log.Println("Starting server on :8087")
	log.Fatal(http.ListenAndServe(":8087", nil))
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check required fields
	if vm.ClientId == 0 || vm.Name == "" || vm.Template == "" || vm.GuestOsId == "" || vm.Cores == 0 || vm.MemorySize == 0 || vm.OperatingSystemDisk.Capacity == 0 || vm.OperatingSystemDisk.StorageProfile == "" || vm.HostingLocation.Id == "" || vm.HostingLocation.Name == "" || vm.HostingLocation.DefaultNetwork == "" || vm.BackupType == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	vmID := uuid.New().String()
	vm.Id = vmID

	vmMutex.Lock()
	vms[vmID] = vm
	vmMutex.Unlock()

	file, err := json.MarshalIndent(vm, "", " ")
	if err != nil {
		http.Error(w, "Error saving VM details", http.StatusInternalServerError)
		return
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s.json", vmID), file, 0644)
	if err != nil {
		http.Error(w, "Error saving VM details", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(map[string]string{"id": vmID})
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

	var allVMs []VirtualMachine

	files, err := ioutil.ReadDir(".")
	if err != nil {
		http.Error(w, "Error reading VM files", http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			data, err := ioutil.ReadFile(file.Name())
			if err != nil {
				http.Error(w, "Error reading VM file", http.StatusInternalServerError)
				return
			}
			var vm VirtualMachine
			err = json.Unmarshal(data, &vm)
			if err != nil {
				http.Error(w, "Error parsing VM file", http.StatusInternalServerError)
				return
			}
			allVMs = append(allVMs, vm)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allVMs)
}

func authenticate(r *http.Request) bool {
	apiKeyHeader := r.Header.Get("Authorization")
	userHeader := r.Header.Get("x-mcs-user")

	return apiKeyHeader == fmt.Sprintf("apiKey %s", apiKey) && userHeader == user
}
