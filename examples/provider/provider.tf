terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

variable "rackspace_spot_token" {
  description = "Rackspace Spot authentication token"
  type        = string
  sensitive   = true
}

provider "spot" {
  token = var.rackspace_spot_token
}
