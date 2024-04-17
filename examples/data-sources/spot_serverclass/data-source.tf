data "spot_serverclass" "example" {
  name = "gp.vs1.medium-dfw"
}

output "serverclass" {
  value = data.spot_serverclass.example
}