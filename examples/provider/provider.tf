terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

provider "spot" {
  # overrides environment variables
  rxtspot_token = "<token>"
}
