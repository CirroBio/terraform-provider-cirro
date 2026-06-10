//go:build tools

package tools

import (
	// tfplugindocs generates provider documentation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
