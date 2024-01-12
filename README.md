Terraform Provider for Rackspace Spot
=======================================

Usage
-------

1. Refer `_examples/simple-cloudspace/main.tf` for usage. You can also refer docs in `./docs` directory.
2. Create `_examples/simple-cloudspace/terraform.tfvars` with values of `apiserver`.
3. Run `terraform apply` command in the `simple-cloudspace` directory.

Development
------------

## Initial Setup

1. Add `dev_overrides` section as the following in the `~/.terraformrc`

```hcl
    provider_installation {
      dev_overrides {
        # Replace with your own GOBIN path. Default is $GOPATH/bin
        "ngpc.rxt.io/rackerlabs/rxtspot" = "/home/<your-username>/go/bin"
      }
      direct {}
    }
```

2. Ensure that `~/go/bin` is in path.
3. Run `make dependencies check-versions`

## Adding new resources

1. Add new data-sources, resources in `provider_code_spec.json` and make corresponding entries in `generator_config.yml`. Refer `openapi-spec/spot-api-3.0.json` to learn about attributes supported by resources and data-sources.
2. Run `make generate` to generate schema corresponding to resources and run `make scaffold-ds` or `scaffold-rs` to scaffold code to interact with resource.
3. Implement Create(), Read(), Update(), Delete() functions from scaffolded go file.
4. Run `make install`
5. Use examples/xx to test provider code.
6. Use `TF_LOG=TRACE terraform plan` to see logs in the output.
