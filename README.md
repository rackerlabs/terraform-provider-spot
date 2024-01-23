Terraform Provider for Rackspace Spot
=======================================

Usage
-------

1. Add a `dev_overrides` section in your `~/.terraformrc` file as follows:

```terraform
    provider_installation {
      dev_overrides {
        # Replace with your own GOBIN path. Default is $GOPATH/bin
        "ngpc.rxt.io/rackerlabs/spot" = "/home/<your-username>/go/bin"
      }
      direct {}
    }
```

2. Ensure that `~/go/bin` is included in your system's PATH.
3. Run `make check-versions` to verify that Go and Terraform are installed.
4. Run `make install` to install the provider.
5. Refer to [`_examples/simple-cloudspace/main.tf`](/_examples/simple-cloudspace/main.tf) for the supported syntax. You can also consult the [documentation](/docs/index.md).
6. Log on to the [Rackspace Spot dashboard](https://spot.rackspace.com) and create an organization or select an existing organization.
7. While being logged in, inspect the network calls to the `https://spot.rackspace.com/apis` and copy the token from Authorization request header.
8. Store the token in a local file.
9. Set an environment variable `RXTSPOT_TOKEN_FILE=/path/to/the/token/file`.
10. If you are testing in a non-production environment, set the `NGPC_APISERVER` environment variable.
11. Create [`_examples/simple-cloudspace/terraform.tfvars`](/_examples/simple-cloudspace/terraform.tfvars) with values of your `organization` and `cloudspace_name`.
12. Run `make apply` command that runs `terraform apply`.
13. Run `make destroy` to cleanup.


Adding new resources
----------------------

1. Add new data sources or resources in `provider_code_spec.json`.
2. Make corresponding entries in `generator_config.yml`.
3. Refer to the [openapi-spec](/openapi-spec/spot-api-3.0.json) to understand the attributes supported by these resources and data sources.
4. Run `make generate` to create schemas corresponding to resources and data sources. The generated files, ending with `_gen.go`, should not be edited manually.
5. Run `NAME=<resource-name> make scaffold-rs` or `scaffold-ds` to scaffold code for interacting with resources and data sources.
6. Implement the necessary functions in the scaffolded Go files, located at `internal/provider/<resource-name>_resource.go`.
7. Run `make install`.
8. To view detailed logs during a Terraform execution, use `TF_LOG=TRACE terraform plan`.


Links
------

- https://developer.hashicorp.com/terraform/plugin/framework
- https://developer.hashicorp.com/terraform/plugin/code-generation
- https://developer.hashicorp.com/terraform/plugin/code-generation/framework-generator