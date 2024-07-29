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
