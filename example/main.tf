variable "api_key" {
  description = "Softsource vBridge API Key"
  type        = string
  sensitive   = true
}

variable "api_user_email" {
  description = "Softsource vBridge user email address"
  type        = string
  sensitive   = true
}

variable "client_id" {
  description = "Softsource vBridge client id"
  type        = number
  sensitive   = true
}

terraform {
  required_providers {
    vbridge = {
      version = "~> 1.0.1"
      source  = "durankeeley.com/vbridge/vbridge-vm"
    }
  }
}

provider "vbridge" {
  api_url    = "https://api.mycloudspace.co.nz"
  api_key    = "${var.api_key}"
  user_email = "${var.api_user_email}"
}

resource "vm_provision" "example" {
  provider = vbridge
  client_id                          = var.client_id
  name                               = "terraformvm"
  template                           = "Windows2022_Standard_30GB"
  guest_os_id                        = "windows2019srv_64Guest"
  cores                              = 2
  memory_size                        = 6
  operating_system_disk_capacity     = 30
  operating_system_disk_storage_profile = "vStorageT1"
  iso_file = ""
  quote_item = {}
  hosting_location_id             = "vcchcres"
  hosting_location_name           = "Christchurch"
  hosting_location_default_network = "CHC-CUST-SDC-WAN"
  backup_type                     = "vBackup"
}
