Terraform Provider for Rackspace Spot
=====================================

The Rackspace Spot Terraform Provider for creating and managing cloudspaces, on [Rackspace Spot](https://spot.rackspace.com/).

The provider is available on the terraform registry at [rackerlabs/spot](https://registry.terraform.io/providers/rackerlabs/spot/latest). You can find the instructions to use the provider in the [official documentation](https://registry.terraform.io/providers/rackerlabs/spot/latest/docs) or in the [docs](./docs/index.md) directory.

## Development Setup

To set up your development environment, follow these steps:
1. Create a new file called `.terraformrc` in your home directory if it doesn't already exist.
2. Add the following `dev_overrides` block to your `.terraformrc` file:
```terraform
provider_installation {
  dev_overrides {
    # Replace the path with the location where the provider binary is installed on your system.
    "rackspace/spot" = "/home/<your-username>/go/bin"
  }
  direct {}
}
```
3. You don't need to run `terraform init` after adding the `dev_overrides` block to the `.terraformrc` file. Terraform automatically uses the development version of the provider when you run `terraform apply` or `terraform plan`.

## Building the Provider

### Requirements

- [Terraform](https://www.terraform.io/downloads.html)
- [Go](https://golang.org/doc/install) to build the provider plugin
- [Visual Studio Code](https://code.visualstudio.com/download) (optional, but recommended)
- [Make](https://www.gnu.org/software/make/) for running the Makefile
- [GoReleaser](https://goreleaser.com/install/) for creating releases

### Development Workflow

```shell
# Build the provider and install its binary in GOBIN path, /home/<your-username>/go/bin
make install

# Add new resource/data source in the `provider_code_spec.json`. Refer existing resource/data-source.
code provider_code_spec.json

# Use the following command to generate corresponding go types
make generate-code

# Scaffold code for a new resource or data source
NAME=newresource make scaffold-rs

# Modify the scaffolded code to implement the resource or data source
code internal/provider/newresource_resource.go

# Add documentation for the new resource or data source, use templates for attributes and examples. Refer existing templates.
code templates/resources/newresource.md.tmpl

# Generate the documentation for terraform registry
make generate
```

### Debugging with Visual Studio Code

1. Open the [`.vscode/launch.json`](/.vscode/launch.json) file and update the value of the `NGPC_APISERVER` variable.
2. Set breakpoints and start the debugging session in Visual Studio Code.
3. Copy the value of the `TF_REATTACH_PROVIDERS` from the *DEBUG CONSOLE* tab in Visual Studio Code.
4. Open a terminal and set the `TF_REATTACH_PROVIDERS` environment variable to the copied value.
5. In the terminal, run `terraform apply` or `terraform plan` to trigger the provider execution and hit the breakpoints.

## Contributing

1. Clone this repository.
2. Make any desired changes in your cloned repository. When you are ready to send those changes to us, push your changes to an upstream branch and [create a pull request](https://help.github.com/articles/creating-a-pull-request/).
3. After your pull request has been approved, it will be merged into the repository.
