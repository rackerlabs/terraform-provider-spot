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
  cloudspace_name    = "name-of-the-cloudspace"
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
  wait_until_ready   = true
  # deployment_type field is optional and the default is "gen1".
  # Supported values: gen1, gen2
  deployment_type = "gen1"
}

# Creates a spot node pool with an autoscaling pool of 3-8 servers of class gp.vs1.large-dfw.
resource "spot_spotnodepool" "autoscaling-bid" {
  cloudspace_name = resource.spot_cloudspace.example.cloudspace_name
  server_class    = "gp.vs1.large-dfw"
  bid_price       = 0.012

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
