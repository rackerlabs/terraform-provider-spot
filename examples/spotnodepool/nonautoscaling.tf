resource "spot_spotnodepool" "example" {
  cloudspace_name      = "example"
  server_class         = "gp.vs1.small-dfw"
  bid_price            = 0.002
  desired_server_count = 2
}