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

# Example of cloudspace resource.
resource "spot_cloudspace" "example" {
  cloudspace_name = "name-of-the-cloudspace"
  # You can find the available region names in the `regions` data source.
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
  wait_until_ready   = true
}

# Creates a spot node pool with an autoscaling pool of 3-8 servers of class gp.vs1.large-dfw.
resource "spot_spotnodepool" "autoscaling-bid" {
  cloudspace_name = resource.spot_cloudspace.example.cloudspace_name
  # You can find the available server classes in the `serverclasses` data source.
  server_class = "gp.vs1.large-dfw"
  bid_price    = 0.012

  autoscaling = {
    min_nodes = 3
    max_nodes = 8
  }
}

data "spot_kubeconfig" "example" {
  cloudspace_name = resource.spot_cloudspace.example.name
}

output "kubeconfig" {
  value = data.spot_kubeconfig.example.raw
}
