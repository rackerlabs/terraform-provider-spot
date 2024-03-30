---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spot_spotnodepool Resource - Rackspace Spot"
subcategory: ""
description: |-
  
---

# spot_spotnodepool Resource

The `spotnodepool` resource is designed to manage Spot Node Pools within a specified cloudspace. Users have the flexibility to create multiple spot node pools, each uniquely configured according to their requirements. These pools are then associated with a specific cloudspace using the cloudspace_name attribute. This setup allows for efficient allocation and management of resources in a cloud environment, enabling users to optimize their cloud infrastructure based on varying workloads and demands.

## Example Usage

```terraform
# Creates a spot node pool with a two servers of class gp.vs1.small-dfw.
resource "spot_spotnodepool" "example" {
  cloudspace_name = "example"
  server_class    = "gp.vs1.small-dfw"
  bid_price       = 0.002
  autoscaling = {
    min_nodes = 2
    max_nodes = 4
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `bid_price` (Number) The bid price for the server in USD, rounded to three decimal places.
- `cloudspace_name` (String) The name of the cloudspace.
- `server_class` (String) The class of servers to use for the node pool.

### Optional

- `autoscaling` (Attributes) Scales the nodes in a cluster based on usage. This block should be omitted to disable autoscaling. (see [below for nested schema](#nestedatt--autoscaling))
- `desired_server_count` (Number) The desired number of servers in the node pool. Should be removed if autoscaling is enabled.

### Read-Only

- `id` (String) The id of the cloudspace.
- `last_updated` (String) The last time the spotnodepool was updated.
- `resource_version` (String) The version of the resource known to local state. This is used to determine if the resource is modified outside of terraform.

<a id="nestedatt--autoscaling"></a>
### Nested Schema for `autoscaling`

Optional:

- `max_nodes` (Number) The maximum number of nodes in the node pool.
- `min_nodes` (Number) The minimum number of nodes in the node pool.

### List of available server classes

| Region          | Name              | Category       | CPU | Memory  |
|-----------------|-------------------|----------------|-----|---------|
| us-central-dfw-1| ch.vs1.2xlarge-dfw| Compute Heavy  | 16  | 30GB    |
| us-central-dfw-1| ch.vs1.large-dfw  | Compute Heavy  | 4   | 7.5GB   |
| us-central-dfw-1| ch.vs1.medium-dfw | Compute Heavy  | 2   | 3.75GB  |
| us-central-dfw-1| ch.vs1.xlarge-dfw | Compute Heavy  | 8   | 15GB    |
| us-central-dfw-1| gp.vs1.2xlarge-dfw| General Purpose| 16  | 60GB    |
| us-central-dfw-1| gp.vs1.large-dfw  | General Purpose| 4   | 15GB    |
| us-central-dfw-1| gp.vs1.medium-dfw | General Purpose| 2   | 3.75GB  |
| us-central-dfw-1| gp.vs1.small-dfw  | General Purpose| 1   | 1GB     |
| us-central-dfw-1| gp.vs1.xlarge-dfw | General Purpose| 8   | 30GB    |
| us-central-dfw-1| mh.vs1.2xlarge-dfw| Memory Heavy   | 16  | 120GB   |
| us-central-dfw-1| mh.vs1.large-dfw  | Memory Heavy   | 4   | 30GB    |
| us-central-dfw-1| mh.vs1.medium-dfw | Memory Heavy   | 2   | 15GB    |
| us-central-dfw-1| mh.vs1.xlarge-dfw | Memory Heavy   | 8   | 60GB    |
| us-east-iad-1   | ch.vs1.2xlarge-iad| Compute Heavy  | 16  | 30GB    |
| us-east-iad-1   | ch.vs1.large-iad  | Compute Heavy  | 4   | 7.5GB   |
| us-east-iad-1   | ch.vs1.medium-iad | Compute Heavy  | 2   | 3.75GB  |
| us-east-iad-1   | ch.vs1.xlarge-iad | Compute Heavy  | 8   | 15GB    |
| us-east-iad-1   | gp.vs1.2xlarge-iad| General Purpose| 16  | 60GB    |
| us-east-iad-1   | gp.vs1.large-iad  | General Purpose| 4   | 15GB    |
| us-east-iad-1   | gp.vs1.medium-iad | General Purpose| 2   | 3.75GB  |
| us-east-iad-1   | gp.vs1.small-iad  | General Purpose| 1   | 1GB     |
| us-east-iad-1   | gp.vs1.xlarge-iad | General Purpose| 8   | 30GB    |
| us-east-iad-1   | mh.vs1.2xlarge-iad| Memory Heavy   | 16  | 120GB   |
| us-east-iad-1   | mh.vs1.large-iad  | Memory Heavy   | 4   | 30GB    |
| us-east-iad-1   | mh.vs1.medium-iad | Memory Heavy   | 2   | 15GB    |
| us-east-iad-1   | mh.vs1.xlarge-iad | Memory Heavy   | 8   | 60GB    |


## Import

Import is supported using the following syntax:

```shell
# A spotnodepool can be imported by specifying its id.
# The id is the organization id(namespace) followed by a slash, followed by the name of the spotnodepool.
terraform import spot_spotnodepool.example org-yxrstzzs6qqokjva/c126b90d-00d1-48fb-92ae-b8c88e27f511
```