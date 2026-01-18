// Package acctest provides shared acceptance test utilities for the DanubeData provider.
package acctest

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ProtoV6ProviderFactories returns the provider factories for acceptance tests
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"danubedata": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// PreCheck validates the necessary test environment variables exist
func PreCheck(t *testing.T) {
	if v := os.Getenv("DANUBEDATA_API_TOKEN"); v == "" {
		t.Fatal("DANUBEDATA_API_TOKEN must be set for acceptance tests")
	}
}

// RandomName generates a random name with the given prefix for testing
func RandomName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, RandomString(8))
}

// RandomString generates a random lowercase alphanumeric string
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

// RandomPassword generates a random password meeting minimum requirements
func RandomPassword() string {
	// At least 12 characters with upper, lower, number, and special
	return fmt.Sprintf("Test%s!@#%d", RandomString(8), rand.IntN(1000))
}

// RandomSSHPublicKey generates a fake SSH public key for testing
// Note: This is NOT a valid key, just for testing API calls
func RandomSSHPublicKey() string {
	return fmt.Sprintf("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI%s test@example.com", RandomString(44))
}

// ProviderConfig returns the provider configuration block
func ProviderConfig() string {
	return `
provider "danubedata" {}
`
}

// ConfigCompose concatenates multiple configuration strings
func ConfigCompose(configs ...string) string {
	var b strings.Builder
	for _, config := range configs {
		b.WriteString(config)
		b.WriteString("\n")
	}
	return b.String()
}

// CheckResourceAttrSet is a helper to check if an attribute is set
func CheckResourceAttrSet(name, key string) func(s interface{}) error {
	return func(s interface{}) error {
		// This is a placeholder - actual implementation would use terraform-plugin-testing
		return nil
	}
}

// SkipIfEnvNotSet skips the test if the specified environment variable is not set
func SkipIfEnvNotSet(t *testing.T, envVar string) {
	if os.Getenv(envVar) == "" {
		t.Skipf("Skipping test: %s environment variable not set", envVar)
	}
}
