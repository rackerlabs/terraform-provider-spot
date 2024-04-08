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

variable "region" {
  description = "The region in which the cloudspace is created"
  type        = string
  default     = "us-central-dfw-1"
}

provider "spot" {}

# Cloudspace resource with default configuration.
resource "spot_cloudspace" "example" {
  cloudspace_name = var.cloudspace_name
  region          = var.region
}

# Creates a spot node pool with two servers of class gp.vs1.medium-dfw.
resource "spot_spotnodepool" "non-autoscaling-bid" {
  cloudspace_name      = resource.spot_cloudspace.example.cloudspace_name
  server_class         = "gp.vs1.medium-dfw"
  bid_price            = 0.007
  desired_server_count = 2
}

data "spot_kubeconfig" "example" {
  id = resource.spot_cloudspace.example.id
}

output "kubeconfig" {
  value = data.spot_kubeconfig.example.raw
}

# Save the kubeconfig to a local file.
resource "local_file" "kubeconfig" {
  depends_on = [data.spot_kubeconfig.example]
  count      = 1
  content    = data.spot_kubeconfig.example.raw
  filename   = "${path.root}/kubeconfig"
}