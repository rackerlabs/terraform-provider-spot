# Find all regions in the USA
data "spot_regions" "usa" {
  filters = [
    {
      name   = "country",
      values = ["USA"]
    }
  ]
}

output "region_test" {
  value = data.spot_regions.usa.names
}