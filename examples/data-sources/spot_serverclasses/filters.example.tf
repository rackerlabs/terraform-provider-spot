# Find serverclasses that are from the category compute heavy
data "spot_serverclasses" "all" {
  filters = [
    {
      name   = "category"
      values = ["Compute Heavy"]
    }
  ]
}

output "names" {
  value = data.spot_serverclasses.all.names
}