---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spot_serverclass Data Source - terraform-provider-spot"
subcategory: ""
description: |-
  
---

# spot_serverclass (Data Source)



## Example Usage

```terraform
data "spot_serverclass" "example" {
  name = "gp.vs1.medium-dfw"
}

output "serverclass" {
  value = data.spot_serverclass.example
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the server class

### Read-Only

- `availability` (String) Describes the serverclass availability status
- `category` (String) Describes the serverclass category
- `display_name` (String) Specifies the human-readable name to use
- `flavor_type` (String) Describes whether it is a VM or bare metal. This determines certain capabilities like nested virtualization
- `on_demand_pricing` (Attributes) (see [below for nested schema](#nestedatt--on_demand_pricing))
- `region` (String) Specifies the region where the servers belonging to this ServerClass resides in
- `resources` (Attributes) (see [below for nested schema](#nestedatt--resources))
- `serverclass_provider` (Attributes) (see [below for nested schema](#nestedatt--serverclass_provider))
- `status` (Attributes) (see [below for nested schema](#nestedatt--status))

<a id="nestedatt--on_demand_pricing"></a>
### Nested Schema for `on_demand_pricing`

Read-Only:

- `cost` (String) Describes the USD cost of this type of servers. If pricing is localized, this can be used as the base factor
- `interval` (String) Indicates the interval used for the pricing


<a id="nestedatt--resources"></a>
### Nested Schema for `resources`

Read-Only:

- `cpu` (String)
- `memory` (String)


<a id="nestedatt--serverclass_provider"></a>
### Nested Schema for `serverclass_provider`

Read-Only:

- `flavor_id` (String) Name of the flavor
- `provider_type` (String) Actual infrastructure backing the server class


<a id="nestedatt--status"></a>
### Nested Schema for `status`

Read-Only:

- `available` (Number) how many servers of this class are currently in use
- `capacity` (Number) how many servers of this class are currently in use
- `last_auction` (Number) how many servers of this class are currently in use
- `reserved` (Number) how many servers of this class are currently in use
- `spot_pricing` (Attributes) (see [below for nested schema](#nestedatt--status--spot_pricing))

<a id="nestedatt--status--spot_pricing"></a>
### Nested Schema for `status.spot_pricing`

Read-Only:

- `hammer_price_per_hour` (String)
- `market_price_per_hour` (String)
