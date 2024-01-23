//go:build tools
// +build tools

package tools

import (
	_ "github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework"
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
