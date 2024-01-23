terraform {
  required_providers {
    spot = {
      source = "ngpc.rxt.io/rackerlabs/spot"
    }
  }
}

provider "spot" {}

resource "spot_cloudspaces" "example" {}
