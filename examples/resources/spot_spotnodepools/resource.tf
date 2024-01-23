# Creates a spot node pool with a two servers of class gp.vs1.small-dfw.
resource "spot_spotnodepools" "example" {
  cloudspace_name      = "example"
  organization         = "my-org"
  server_class         = "gp.vs1.small-dfw"
  bid_price            = "0.002"
  desired_server_count = 2
  autoscaling = {
    enabled   = true
    min_nodes = 2
    max_nodes = 4
  }
}
