# Steps followed to Build a Terraform Provider from OpenAPI Spec

1. To generate something called "provider_code_spec" from openapi spec we need "openapi-generator". Read more here: [Terraform OpenAPI Generator Documentation](https://developer.hashicorp.com/terraform/plugin/code-generation/openapi-generator). We will figure out what is "provider_code_spec" in further steps.

    Install openapi generator:
    ```
    go install github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi@latest
    ```

    Download example OpenAPI spec from [here](https://spot.rackspace.com/docs/rxt_spot_api). Click OAS 2 button on the right-top. Note that this is OpenAPI spec 2.0, which is earlier format called swagger. That is not supported by the "tfplugingen-openapi-generator". Convert it to OAPI 3.0.x using online tools and name it `spot-api-3.0.json`.

    Write `generator_config.yml`, that classifies resources from `spot-api-3.0.json` into provider, resources and data-sources.

    ```
    tfplugingen-openapi generate \
      --config ./generator_config.yml \
      --output ./provider_code_spec.json \
      ./openapi.json
    ```
    This generates `provider_code_spec.json` in the current directory. The `provider_code_spec.json` can be used by "framework generator" in next steps to generate actual go-code for our terraform provider.

2. Before generating code let's set up a go-project. Learn more: [Terraform Workflow Example](https://developer.hashicorp.com/terraform/plugin/code-generation/workflow-example)

    ```
    go mod init terraform-provider-spot
    touch main.go
    mkdir -p internal/provider # this is the directory where the framework generator will generate go-code
    ```

    Install framework generator:
    ```
    go install github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@latest
    ```

    Understand `provider_code_spec.json` specifications here: [Terraform Code Generation Specification](https://developer.hashicorp.com/terraform/plugin/code-generation/specification). It is the only input to our code generator; hence carefully review it.

    Generate "Terraform Provider code":
    ```
    tfplugingen-framework generate all --input ./provider_code_spec.json --output internal/provider
    ```

    Generate starter code that initializes provider (boilerplate):
    ```
    tfplugingen-framework scaffold provider --name spot --output-dir ./internal/provider
    ```
    This command generates `internal/provider.go` which contains the implementation of `provider.Provider` interface.

    `provider.Provider` is a Terraform provider's interface, which we should implement.

3. Create `main.go` to hookup the empty Petstore provider to a provider server. Learn more about provider servers: [Terraform Provider Servers](https://developer.hashicorp.com/terraform/plugin/framework/provider-servers)

    ```go
    package main

    import (
        "context"
        "log"
        "github.com/rackerlabs/terraform-provider-spot/internal/provider"
        "github.com/hashicorp/terraform-plugin-framework/providerserver"
    )

    func main() {
        opts := providerserver.ServeOpts{
            Address: "ngpc.rxt.io/rackerlabs/spot",
        }

        err := providerserver.Serve(context.Background(), provider.New(), opts)
        if err != nil {
            log.Fatal(err.Error())
        }
    }
    ```

    Build the provider:
    ```
    go mod tidy
    go install . # Build and copies binary to GOBIN, but this binary can't be run as a program. make sure GOBIN is in PATH
    ```

4. Inform the Terraform CLI where to find the locally built Petstore provider. By default, the Terraform CLI reads `~/.terraformrc`. Learn more: [Terraform CLI Development Overrides](https://developer.hashicorp.com/terraform/plugin/debugging#terraform-cli-development-overrides). Hence, add the following in the `~/.terraformrc`:

    ```hcl
    provider_installation {
      dev_overrides {
        # Replace with your own GOBIN path. Default is $GOPATH/bin
        "ngpc.rxt.io/rackerlabs/spot" = "/home/nilesh/go/bin"
      }
      direct {}
    }
    ```

5. Also, create `examples/simple-cloudspace/main.tf`:

    ```hcl
    terraform {
      required_providers {
        petstore = {
          source = "ngpc.rxt.io/rackerlabs/spot"
        }
      }
    }

    provider "spot" {}
    ```

6. Scaffold resources, data-sources:

    ```
    tfplugingen-framework scaffold resource --name cloudspace --output-dir ./internal/provider # Generates ./internal/cloudspace_data_source.go
    tfplugingen-framework scaffold data-source --name cloudspace --output-dir ./internal/provider # Generates ./internal/cloudspace_data_source.go
    tfplugingen-framework scaffold data-source --name cloudspaces --output-dir ./internal/provider
    tfplugingen-framework scaffold data-source --name regions --output-dir ./internal/provider
    ```

    This generates `internal/xxx_resource.go`.

7. Link generated code with scaffolded code:

    - Add `NewXXResource` functions in `provider.go/Resources()` and `NewXXDataSource` function is `DataSources()`.
    - Delete all structs named `xxModel` from `xx_resource.go` and replace all references to `resource_xx.xx_resource.gen.go`.
    - Also replace `(r *cloudspaceResource) Schema()` with `resp.Schema = resource_cloudspace.CloudspaceResourceSchema(ctx)`
    - Add `Configure()` function in provider and resources. This function should initialize client, ngpc client in this case.


References
-------------

- https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider
- https://github.com/hashicorp/terraform-plugin-codegen-spec/blob/main/spec/v0.1/schema.json
- https://github.com/hashicorp/terraform-plugin-codegen-openapi/tree/main/internal/cmd/testdata/scaleway
- https://github.com/hashicorp/terraform-provider-kubernetes/blob/main/kubernetes/resource_kubernetes_deployment_v1.go#L246
- https://developer.hashicorp.com/terraform/plugin/best-practices/
- hashicorp-provider-design-principles#resources-should-represent-a-single-api-object
- https://github.com/hashicorp/terraform-provider-hashicups/blob/provider-configure/internal/provider/example_resource.go
- https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-resource-update
