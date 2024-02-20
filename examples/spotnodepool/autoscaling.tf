# Creates a spot node pool with a two servers of class gp.vs1.small-dfw and autoscaling enabled.
# If load increases, the number of nodes will increase up to 4 from the minimum of 2.
resource "spot_spotnodepool" "example" {
  cloudspace_name = "example"
  server_class    = "gp.vs1.small-dfw"
  bid_price       = 0.002
  autoscaling = {
    min_nodes = 2
    max_nodes = 4
  }
}
