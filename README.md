Terraform Provider for Rackspace Spot
=======================================

The Rackspace Spot Terraform Provider is a useful for creating and managing multiple cloudspaces, on Rackspace's robust infrastructure.

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

1. **Log in to the [Rackspace Spot Console](https://spot.rackspace.com):**
   - If you don't have an account, sign up for one.
   - Log in with your credentials or SSO.

2. **Select an Organization:**
   - If you haven't created an organization, create one by following the instructions.

3. **Access the Spot Dashboard:**
   - After creating or selecting an organization, you should land on the Spot dashboard.

4. **Get the Access Token:**
   - Navigate to the **Terraform** menu under **API Access** on the left pane.
   - Copy the **Access Token** provided on that page.
   - **Important:** Treat the access token as sensitive information. Avoid sharing it publicly.

5. **Set the copied token as an environment variable:**

     ```bash
     export RXTSPOT_TOKEN=<your_access_token>
     ```
6. **Create your first Cloudspace**

    ```terraform
    terraform {
      required_providers {
         spot = {
            source = "rackerlabs/spot"
         }
      }
   }

    # Provider does not require any additional configuration 
    # except the RXTSPOT_TOKEN or RXTSPOT_TOKEN_FILE environment variable
    provider "spot" {}

    # Example of cloudspace resource.
    resource "spot_cloudspace" "example" {
      cloudspace_name    = "example"
      organization       = "my-org"
      region             = "us-central-dfw-1"
      hacontrol_plane    = false
      preemption_webhook = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
    }

    # Creates a spot node pool with a two servers of class gp.vs1.small-dfw.
    resource "spot_spotnodepools" "example" {
      cloudspace_name      = "example"
      organization         = "my-org"
      server_class         = "gp.vs1.small-dfw"
      bid_price            = "0.002"
      desired_server_count = 2
      autoscaling = {
        enabled   = true
        min_nodes = 2
        max_nodes = 4
      }
    }
    ```
