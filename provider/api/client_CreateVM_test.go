package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVM(t *testing.T) {
	// Given
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := []map[string]interface{}{
			{
				"id":              12345,
				"name":            "test-vm-1",
				"hostingLocation": "Christchurch",
			},
			{
				"id":              12346,
				"name":            "test-vm-2",
				"hostingLocation": "Wellington",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	vm := VirtualMachine{
		ClientId:   123,
		Name:       "test-vm-1",
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
	assert.Equal(t, "12345", result, "VM ID mismatch")
}
