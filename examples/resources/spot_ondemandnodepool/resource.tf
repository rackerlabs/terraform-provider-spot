# Creates a ondemand node pool with a two servers of class gp.vs1.small-dfw.
resource "spot_ondemandnodepool" "example" {
  cloudspace_name      = "example"
  server_class         = "gp.vs1.small-dfw"
  desired_server_count = 2
}
