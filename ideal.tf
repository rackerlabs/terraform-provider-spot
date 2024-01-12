

provider "rackspacespot" {
  apiserver  = "https://ngpc-staging-3.platform9.horse"
  auth_token = ""
}

resource "cloudspace" "my-cloudspace" {
  name               = "my-cloudspace"
  organization       = "default"
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  k8s_version        = "1.18"
  preemption_webhook = ""

  node_pool {
    server_class         = "gp.vs1.small-dfw"
    desired_server_count = 1
    bid_price            = "2.008"
  }

  node_pool {
    server_class         = "gp.vs1.general-dfw"
    desired_server_count = 1
    bid_price            = "0.001"
  }
}
