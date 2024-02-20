terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

variable "cloudspace_name" {
  type = string
}

provider "spot" {}

resource "spot_cloudspace" "my-cloudspace" {
  cloudspace_name    = var.cloudspace_name
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
}

resource "spot_spotnodepool" "small-nodes" {
  cloudspace_name      = spot_cloudspace.my-cloudspace.cloudspace_name
  server_class         = "gp.vs1.small-dfw-xyz"
  bid_price            = 2.001
  autoscaling = {
    min_nodes = 2
    max_nodes = 4
  }
}

resource "spot_spotnodepool" "medium-nodes" {
  cloudspace_name      = spot_cloudspace.my-cloudspace.cloudspace_name
  server_class         = "gp.vs1.medium-dfw"
  bid_price            = 1.012
  desired_server_count = 2
}

data "spot_cloudspace" "my-cloudspace" {
  id = resource.spot_cloudspace.my-cloudspace.id
}

output "kubeconfig" {
  value = data.spot_cloudspace.my-cloudspace.kubeconfig
}
