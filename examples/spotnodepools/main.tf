terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

provider "spot" {}

resource "spot_spotnodepools" "example" {
  cloudspace_name      = "example"
  organization         = "my-org"
  server_class         = "gp.vs1.small-dfw"
  bid_price            = "0.002"
  desired_server_count = 2
}

resource "spot_cloudspace" "my-cloudspace" {
  cloudspace_name    = "example"
  organization       = "my-org"
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = "https://hooks.slack.com/services/TXX"
}
