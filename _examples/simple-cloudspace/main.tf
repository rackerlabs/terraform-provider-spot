terraform {
  required_providers {
    spot = {
      source = "ngpc.rxt.io/rackerlabs/spot"
    }
  }
}

variable "cloudspace_name" {
  type = string
}

variable "organization" {
  type = string
}

provider "spot" {}

resource "spot_cloudspace" "my-cloudspace" {
  cloudspace_name    = var.cloudspace_name
  organization       = var.organization
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = ""
}

resource "spot_spotnodepools" "small-nodes" {
  cloudspace_name      = spot_cloudspace.my-cloudspace.cloudspace_name
  organization         = spot_cloudspace.my-cloudspace.organization
  server_class         = "gp.vs1.small-dfw"
  bid_price            = "10.002"
  desired_server_count = 2
  autoscaling = {
    enabled   = true
    min_nodes = 2
    max_nodes = 4
  }
}

resource "spot_spotnodepools" "medium-nodes" {
  cloudspace_name      = spot_cloudspace.my-cloudspace.cloudspace_name
  organization         = spot_cloudspace.my-cloudspace.organization
  server_class         = "gp.vs1.medium-dfw"
  bid_price            = "0.012"
  desired_server_count = 2
}
