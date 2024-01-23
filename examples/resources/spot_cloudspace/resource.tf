# Example of cloudspace resource.
resource "spot_cloudspace" "example" {
  cloudspace_name    = "example"
  organization       = "my-org"
  region             = "us-central-dfw-1"
  hacontrol_plane    = false
  preemption_webhook = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
}
