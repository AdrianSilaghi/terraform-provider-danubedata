package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during
// acceptance testing. The factory function is called for each Terraform CLI
// command executed during acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"danubedata": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates the necessary test environment variables exist.
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DANUBEDATA_API_TOKEN"); v == "" {
		t.Fatal("DANUBEDATA_API_TOKEN must be set for acceptance tests")
	}
}

// TestProvider_HasResources verifies the provider has the expected resources
func TestProvider_HasResources(t *testing.T) {
	p := New("test")()

	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "danubedata" {
		t.Errorf("expected provider type name 'danubedata', got %s", resp.TypeName)
	}
}

// TestProvider_HasExpectedResourceTypes verifies the provider registers expected resources
func TestProvider_HasExpectedResourceTypes(t *testing.T) {
	p := New("test")()

	resources := p.Resources(context.Background())

	// Verify we have the expected number of resources
	expectedResourceCount := 9 // vps, serverless, cache, database, storage_bucket, storage_access_key, ssh_key, firewall, vps_snapshot
	if len(resources) != expectedResourceCount {
		t.Errorf("expected %d resources, got %d", expectedResourceCount, len(resources))
	}
}

// TestProvider_HasExpectedDataSourceTypes verifies the provider registers expected data sources
func TestProvider_HasExpectedDataSourceTypes(t *testing.T) {
	p := New("test")()

	dataSources := p.DataSources(context.Background())

	// Verify we have the expected number of data sources
	expectedDataSourceCount := 4 // vps_images, cache_providers, database_providers, ssh_keys
	if len(dataSources) != expectedDataSourceCount {
		t.Errorf("expected %d data sources, got %d", expectedDataSourceCount, len(dataSources))
	}
}
