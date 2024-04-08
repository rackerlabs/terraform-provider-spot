data "spot_cloudspace" "example" {
  id = "1d7f9f9b-9e8d-4c8c-a9d5-e6e6f6f6f6f6"
}

# This outputs the current phase of the cloudspace.
output "csphase" {
  value = data.spot_cloudspace.example.phase
}