package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVM(t *testing.T) {
	// Counter to track the number of GetVMByName calls
	var getVMByNameCalls int

	// Given
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle POST /api/Provisioning/VirtualMachine
		if r.Method == "POST" && r.URL.Path == "/api/Provisioning/VirtualMachine" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(``))
			return
		}

		// Handle GET /api/client/virtualresources/{clientId}
		if r.Method == "GET" && r.URL.Path == "/api/client/virtualresources/123" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Simulate VM becoming available after the first call
			getVMByNameCalls++
			var response []map[string]interface{}
			if getVMByNameCalls >= 2 {
				response = []map[string]interface{}{
					{
						"id":              12345,
						"name":            "test-vm-1",
						"hostingLocation": "Christchurch",
					},
					{
						"id":              12346,
						"name":            "test-vm-2",
						"hostingLocation": "Auckland",
					},
				}
			} else {
				response = []map[string]interface{}{
					{
						"id":              12345,
						"name":            "test-vm-1",
						"hostingLocation": "Christchurch",
					},
				}
			}

			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	vm := VirtualMachine{
		ClientId:   123,
		Name:       "test-vm-2",
		Template:   "template-123",
		GuestOsId:  "os-123",
		Cores:      4,
		MemorySize: 8192,
		OperatingSystemDisk: VirtualDisk{
			Capacity: 100,
		},
	}

	client := testClient(mockServer.URL)

	// When
	result, err := client.CreateVM(vm)

	// Then
	assert.NoError(t, err, "expected no error from CreateVM")
	assert.Equal(t, "12346", result, "VM ID mismatch")
	assert.Equal(t, 2, getVMByNameCalls, "expected 1 call to GetVMByName")
}
