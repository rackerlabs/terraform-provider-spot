data "spot_region" "example" {
  name = "us-east-iad-1"
}

# Outputs the human readable name of the region
output "region_name" {
  value = data.spot_region.example.description
}