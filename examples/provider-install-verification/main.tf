terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
  }
}

variable "token" {
  description = "The rackspace spot token"
  type        = string
}

provider "spot" {
  token = var.token
}
