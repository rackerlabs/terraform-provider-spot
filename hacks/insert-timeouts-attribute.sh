#!/bin/bash

# This script is a temporary fix to the https://github.com/hashicorp/terraform-plugin-codegen-framework/issues/143 

FILE="internal/provider/resource_cloudspace/cloudspace_resource_gen.go"

sed -i '/"github.com\/hashicorp\/terraform-plugin-framework\/types"/i\\t"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"' "$FILE"

sed -i '/"wait_until_ready": schema.BoolAttribute{/i\\t\t\t"timeouts": timeouts.Attributes(ctx, timeouts.Opts{\n\t\t\t\tCreate: true,\n\t\t\t}),' "$FILE"

sed -i '/WaitUntilReady[[:space:]]*types\.Bool/i\\tTimeouts timeouts.Value `tfsdk:"timeouts"`' "$FILE"
