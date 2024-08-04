package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	APIUrl    string
	APIKey    string
	UserEmail string
}

func NewClient(apiURL, apiKey, userEmail string) (*Client, error) {
	return &Client{
		APIUrl:    apiURL,
		APIKey:    apiKey,
		UserEmail: userEmail,
	}, nil
}

// Temp Solution
// Convert Capacity from float to int for GetVMDetailedByID
func (vd *VirtualDisk) UnmarshalJSON(data []byte) error {
	type Alias VirtualDisk
	aux := &struct {
		Capacity float64 `json:"capacity"`
		*Alias
	}{
		Alias: (*Alias)(vd),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	vd.Capacity = int(aux.Capacity)
	return nil
}

func (c *Client) CreateVM(vm VirtualMachine) (string, error) {
	jsonData, err := json.Marshal(vm)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/Provisioning/VirtualMachine", c.APIUrl), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("x-mcs-user", c.UserEmail)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error response from API: %s - %s", resp.Status, string(bodyBytes))
	}

	time.Sleep(35 * time.Second)

	vmID, err := c.GetVMByName(vm.Name, vm.ClientId)
	if err != nil {
		return "", fmt.Errorf("error retrieving VM by name: %w", err)
	}

	return vmID, nil
}

func (c *Client) GetVMByName(vmName string, clientId int) (string, error) {
	url := fmt.Sprintf("%s/api/client/virtualresources/%d", c.APIUrl, clientId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("x-mcs-user", c.UserEmail)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error response from API: %s - %s", resp.Status, string(bodyBytes))
	}

	var vms []struct {
		Id              int    `json:"id"`
		Name            string `json:"name"`
		HostingLocation string `json:"hostingLocation"`
	}

	err = json.NewDecoder(resp.Body).Decode(&vms)
	if err != nil {
		return "", fmt.Errorf("error decoding JSON response: %w", err)
	}

	for _, vm := range vms {
		if vm.Name == vmName {
			return fmt.Sprintf("%d", vm.Id), nil
		}
	}

	return "", fmt.Errorf("VM with name %s not found", vmName)
}

func (c *Client) GetVMDetailedByID(vmID string) (VirtualMachine, error) {
	url := fmt.Sprintf("%s/api/VirtualResource/Detailed/%s", c.APIUrl, vmID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return VirtualMachine{}, fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("x-mcs-user", c.UserEmail)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return VirtualMachine{}, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return VirtualMachine{}, fmt.Errorf("error response from API: %s - %s", resp.Status, string(bodyBytes))
	}

	var temp struct {
		VirtualMachine
		HostingLocation string `json:"hostingLocation"`
	}

	err = json.NewDecoder(resp.Body).Decode(&temp)
	if err != nil {
		return VirtualMachine{}, fmt.Errorf("error decoding JSON response: %w", err)
	}

	vm := temp.VirtualMachine
	vm.HostingLocation = HostingLocation{Name: temp.HostingLocation}

	return vm, nil
}

func (c *Client) PowerOffVM(vmID string) error {
	url := fmt.Sprintf("%s/api/virtualresource/poweroperation", c.APIUrl)
	payload := struct {
		VirtualResourceId string `json:"VirtualResourceId"`
		Operation         string `json:"Operation"`
	}{
		VirtualResourceId: vmID,
		Operation:         "off",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("x-mcs-user", c.UserEmail)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error response from API: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (c *Client) DeleteVM(vmID string, moRef string) error {
	url := fmt.Sprintf("%s/api/virtualresource/delete", c.APIUrl)
	payload := struct {
		VirtualResourceId string `json:"VirtualResourceId"`
		CheckToken        string `json:"CheckToken"`
	}{
		VirtualResourceId: vmID,
		CheckToken:        moRef,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("x-mcs-user", c.UserEmail)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error response from API: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}
func (c *Client) CreateAdditionalDisk(vmID string, disk VirtualDisk) error {
	url := fmt.Sprintf("%s/api/virtualresource/AddDisk", c.APIUrl)
	payload := struct {
		VirtualResourceId string `json:"virtualResourceId"`
		Tier              string `json:"tier"`
		Size              int    `json:"size"`
	}{
		VirtualResourceId: vmID,
		Tier:              disk.StorageProfile,
		Size:              disk.Capacity,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("x-mcs-user", c.UserEmail)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error response from API: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (c *Client) CreateAdditionalDiskWithComparison(vmID string, disk VirtualDisk) (string, error) {
	initialVM, err := c.GetVMDetailedByID(vmID)
	if err != nil {
		return "", fmt.Errorf("error getting VM details before adding disk: %w", err)
	}
	initialDisks := initialVM.Specification.VirtualDisks

	fmt.Println("Creating additional disk...")
	err = c.CreateAdditionalDisk(vmID, disk)
	if err != nil {
		return "", fmt.Errorf("error creating additional disk: %w", err)
	}

	time.Sleep(1 * time.Minute)

	updatedVM, err := c.GetVMDetailedByID(vmID)
	if err != nil {
		return "", fmt.Errorf("error getting VM details after adding disk: %w", err)
	}
	updatedDisks := updatedVM.Specification.VirtualDisks

	newDiskMoRef := findNewDiskMoRef(initialDisks, updatedDisks)
	if newDiskMoRef == "" {
		return "", fmt.Errorf("no new disk was detected")
	}

	fmt.Printf("New disk added with MoRef: %s\n", newDiskMoRef)

	return newDiskMoRef, nil
}

func findNewDiskMoRef(initialDisks, updatedDisks []VirtualDisk) string {
	initialDiskMap := make(map[string]bool)
	for _, disk := range initialDisks {
		initialDiskMap[disk.MoRef] = true
	}

	for _, disk := range updatedDisks {
		if _, found := initialDiskMap[disk.MoRef]; !found {
			return disk.MoRef
		}
	}

	return ""
}

func (c *Client) GetAdditionalDisk(vmID string, diskID string) (*VirtualDisk, error) {
	vm, err := c.GetVMDetailedByID(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM details: %v", err)
	}

	for _, vmDisk := range vm.Specification.VirtualDisks {
		if vmDisk.MoRef == diskID {
			return &vmDisk, nil
		}
	}

	return nil, fmt.Errorf("disk with MoRef %s not found in VM %s", diskID, vmID)
}
