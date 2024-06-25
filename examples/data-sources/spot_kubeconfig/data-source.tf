terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

variable "cloudspace_name" {
  description = "The cloudspace name"
  type        = string
}

variable "token" {
  description = "The rxt spot token"
  type        = string
}

provider "spot" {
  token = var.token
}

data "spot_kubeconfig" "example" {
  cloudspace_name = var.cloudspace_name
}

output "kubeconfig" {
  value = data.spot_kubeconfig.example.raw
}
