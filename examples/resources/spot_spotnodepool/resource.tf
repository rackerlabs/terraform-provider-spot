# Creates a spot node pool with a two servers of class gp.vs1.small-dfw.
resource "spot_spotnodepool" "example" {
  cloudspace_name = "example"
  server_class    = "gp.vs1.small-dfw"
  bid_price       = 0.002
  autoscaling = {
    min_nodes = 2
    max_nodes = 4
  }
}
