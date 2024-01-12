# Example of creation of cloudspace.
resource "rxtspot_cloudspace" "my-cloudspace" {
  cloudspace_name    = "my-cloudspace"
  organization       = "my-org"
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = ""
}
