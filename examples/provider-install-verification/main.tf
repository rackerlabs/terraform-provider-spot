terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

# Set the environemnt variable RXTSPOT_TOKEN to your Spot API token
provider "spot" {}
