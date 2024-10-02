package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
