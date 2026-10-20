package acctest

import (
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGeneratedAcceptanceInputs(t *testing.T) {
	if got := RandomName("terraform"); !regexp.MustCompile(`^terraform-[a-z0-9]{8}$`).MatchString(got) {
		t.Fatalf("invalid resource name %q", got)
	}
	if got := RandomString(0); got != "" {
		t.Fatalf("zero-length string = %q", got)
	}
	if got := RandomPassword(); !regexp.MustCompile(`^Test[a-z0-9]{8}!@#[0-9]{1,3}$`).MatchString(got) {
		t.Fatal("generated password does not satisfy acceptance-test constraints")
	}
	key, comment, _, rest, err := ssh.ParseAuthorizedKey([]byte(RandomSSHPublicKey()))
	if err != nil || key.Type() != "ssh-ed25519" || comment != "test@example.com" || len(rest) != 0 {
		t.Fatalf("generated SSH public key is invalid: %v", err)
	}
	if got := ConfigCompose(ProviderConfig(), `resource "danubedata_ssh_key" "test" {}`); !strings.Contains(got, `provider "danubedata" {}`) || !strings.HasSuffix(got, "{}\n") {
		t.Fatalf("configuration composition = %q", got)
	}
}

func TestAcceptanceEnvironmentGates(t *testing.T) {
	t.Setenv("DANUBEDATA_API_TOKEN", "offline-token")
	PreCheck(t)
	SkipIfEnvNotSet(t, "DANUBEDATA_API_TOKEN")
	t.Setenv("DANUBEDATA_TEST_ABSENT", "")
	t.Run("missing optional environment", func(t *testing.T) {
		SkipIfEnvNotSet(t, "DANUBEDATA_TEST_ABSENT")
		t.Fatal("missing environment variable did not skip")
	})
}
