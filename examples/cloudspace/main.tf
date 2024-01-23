terraform {
  required_providers {
    spot = {
      source = "ngpc.rxt.io/rackerlabs/spot"
    }
  }
}

provider "spot" {}

resource "spot_cloudspace" "my-cloudspace" {
  cloudspace_name    = "example"
  organization       = "my-org"
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = ""
}

resource "spot_spotnodepools" "example" {
  cloudspace_name      = "example"
  organization         = "my-org"
  server_class         = "gp.vs1.small-dfw"
  bid_price            = "0.002"
  desired_server_count = 2
  autoscaling = {
    enabled   = true
    min_nodes = 2
    max_nodes = 4
  }
}
