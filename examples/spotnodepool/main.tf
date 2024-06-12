terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

provider "spot" {}

resource "spot_cloudspace" "my-cloudspace" {
  cloudspace_name    = "example"
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = "https://hooks.slack.com/services/TXX"
  wait_until_ready   = true
}

resource "spot_spotnodepool" "example" {
  cloudspace_name      = "example"
  server_class         = "gp.vs1.small-dfw"
  bid_price            = 0.002
  desired_server_count = 2
}
