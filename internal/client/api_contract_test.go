package client

import (
	_ "embed"
	"encoding/json"
	"testing"
)

//go:embed testdata/api-responses.json
var publishedResponses []byte

func TestDecodePublishedAPIResponses(t *testing.T) {
	var contract struct {
		Responses map[string]json.RawMessage `json:"responses"`
	}
	if err := json.Unmarshal(publishedResponses, &contract); err != nil {
		t.Fatal(err)
	}
	models := map[string]interface{}{
		"VpsInstanceResource":         &VpsInstance{},
		"CacheInstanceResource":       &CacheInstance{},
		"DatabaseInstanceResource":    &DatabaseInstance{},
		"FirewallResource":            &Firewall{},
		"ParameterGroupResource":      &ParameterGroup{},
		"ServerlessContainerResource": &ServerlessContainer{},
		"SshKeyResource":              &SshKey{},
		"StaticSiteResource":          &StaticSite{},
		"StaticSiteDomainResource":    &StaticSiteDomain{},
		"StorageAccessKeyResource":    &StorageAccessKey{},
		"StorageBucketResource":       &StorageBucket{},
		"VpsSnapshotResource":         &VpsSnapshot{},
		"CacheSnapshotResource":       &CacheSnapshot{},
		"DatabaseSnapshotResource":    &DatabaseSnapshot{},
	}
	for name, model := range models {
		t.Run(name, func(t *testing.T) {
			if err := json.Unmarshal(contract.Responses[name], model); err != nil {
				t.Fatalf("published response cannot be decoded: %v", err)
			}
		})
	}
}
