data "spot_cloudspace" "example" {
  id = "mycloudspace"
}

# This outputs the current phase of the cloudspace.
output "csphase" {
  value = data.spot_cloudspace.example.phase
}