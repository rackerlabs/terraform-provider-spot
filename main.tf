terraform {
  required_providers {
    rackspacespot = {
      source = "ngpc.rxt.io/rackspace/rackspacespot"
    }
  }
}

provider "rackspacespot" {
  ngpc_apiserver     = "https://ngpc-staging-3.platform9.horse"
  # auth token should be set as an environment variable NGPC_AUTH_TOKEN
}

resource "rackspacespot_cloudspace" "default" {
  api_version   = "ngpc.rxt.io/v1"
  kind = "CloudSpace"

  metadata = {
    name = "nilest-cs-4"
    namespace = "org-yxrstzzs6qqokjva"
  }

  spec = {
    cloud = "default"
    region = "us-central-dfw-1"
    hacontrol_plane = false
    webhook = ""
  }
}

# resource "rackspacespot_spotnodepool" "default" {
#   api_version   = "ngpc.rxt.io/v1"
#   kind = "SpotNodePool"

#   metadata = {
#     name = "anything"
#     namespace = "org-yxrstzzs6qqokjva"
#   }

#   spec = {
#     bid_price = "2.008"
#     cloud_space = "nilest-cs-3"
#     desired = 3
#     server_class = "gp.vs1.small-dfw"
#   }
# }
