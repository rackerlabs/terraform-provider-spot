data "spot_serverclasses" "all" {
}

output "names" {
  value = data.spot_serverclasses.all.names
}