package shim

import (
	"github.com/hashicorp/terraform-plugin-framework/provider"
	tfprovider "github.com/AdrianSilaghi/terraform-provider-danubedata/internal/provider"
)

// NewProvider returns the DanubeData Terraform provider factory for use with
// the Pulumi Terraform Bridge.
func NewProvider(version string) func() provider.Provider {
	return tfprovider.New(version)
}
