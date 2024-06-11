terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

provider "spot" {}

# Example of cloudspace resource.
resource "spot_cloudspace" "example" {
  cloudspace_name    = "name your cloudspace"
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
}

# Creates a spot node pool with two servers of class gp.vs1.medium-dfw.
resource "spot_spotnodepool" "non-autoscaling-bid" {
  cloudspace_name      = resource.spot_cloudspace.example.cloudspace_name
  server_class         = "gp.vs1.medium-dfw"
  bid_price            = 0.008
  desired_server_count = 2
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

data "spot_cloudspace" "example" {
  id = resource.spot_cloudspace.example.id
}

output "kubeconfig" {
  value = data.spot_cloudspace.example.kubeconfig
}