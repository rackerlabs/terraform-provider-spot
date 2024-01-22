terraform {
  required_providers {
    rxtspot = {
      source = "ngpc.rxt.io/rackerlabs/rxtspot"
    }
  }
}

variable "cloudspace_name" {
  type = string
}

variable "organization" {
  type = string
}

provider "rxtspot" {}

resource "rxtspot_cloudspace" "my-cloudspace" {
  cloudspace_name    = var.cloudspace_name
  organization       = var.organization
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = ""
}

resource "rxtspot_spotnodepools" "small-servers" {
  cloudspace_name      = rxtspot_cloudspace.my-cloudspace.cloudspace_name
  organization         = rxtspot_cloudspace.my-cloudspace.organization
  server_class         = "gp.vs1.small-dfw"
  bid_price            = "0.002"
  desired_server_count = 2
  autoscaling = {
    enabled   = true
    min_nodes = 2
    max_nodes = 4
  }
}

resource "rxtspot_spotnodepools" "general-servers" {
  cloudspace_name      = rxtspot_cloudspace.my-cloudspace.cloudspace_name
  organization         = rxtspot_cloudspace.my-cloudspace.organization
  server_class         = "gp.vs1.small-dfw"
  bid_price            = "0.002"
  desired_server_count = 2
}
