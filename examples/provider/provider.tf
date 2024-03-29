terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

# Provider does not require any additional configuration 
# except the RXTSPOT_TOKEN or RXTSPOT_TOKEN_FILE environment variable
provider "spot" {}
