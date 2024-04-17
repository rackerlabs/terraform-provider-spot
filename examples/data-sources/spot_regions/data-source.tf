data "spot_regions" "available" {
}

output "regions" {
  value = data.spot_regions.available.names
}