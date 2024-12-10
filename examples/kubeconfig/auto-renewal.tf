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
  description = "The spot token"
  type        = string
  sensitive   = true
}

variable "kubeconfig_path" {
  description = "The path to the output kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

provider "spot" {
  token = var.token
}

data "spot_cloudspace" "example" {
  name = var.cloudspace_name
}

data "spot_kubeconfig" "example" {
  cloudspace_name = data.spot_cloudspace.example.name
  depends_on      = [data.spot_cloudspace.example]
}

locals {
  kubeconfig_path = pathexpand(var.kubeconfig_path)
}

# Updates the kubeconfig file with a new token each time `terraform apply` is executed.
resource "local_file" "kubeconfig" {
  depends_on = [data.spot_kubeconfig.example]
  count      = 1
  content    = data.spot_kubeconfig.example.raw
  filename   = local.kubeconfig_path
}