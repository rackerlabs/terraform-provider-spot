---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "Terraform Provider for Rackspace Spot"
subcategory: ""
description: |-
  The rackspace spot terraform provider offers a streamlined solution for creating and managing cloudspaces on Rackspace's robust infrastructure.
---

# Rackspace Spot Provider

The Rackspace Spot Provider is a solution for creating and managing multiple cloudspaces, on Rackspace's robust infrastructure.

## Authenticating with Rackspace Spot

To use this provider, set an authentication token as an environment variable, obtainable via the Rackspace Spot dashboard, https://spot.rackspace.com.

- `RXTSPOT_TOKEN`:  This is the actual token value.
- `RXTSPOT_TOKEN_FILE`: This is the absolute path to the file containing the token value.

```bash
export RXTSPOT_TOKEN=<rackspace-spot-token>
# or
export RXTSPOT_TOKEN_FILE=/path/to/token/file
```

## Example Usage

```terraform
# Provider does not require any additional configuration 
# except the RXTSPOT_TOKEN or RXTSPOT_TOKEN_FILE environment variable
provider "spot" {}
```