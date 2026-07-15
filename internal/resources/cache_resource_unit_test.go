package resources

import (
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMapCacheToState_PreservesParameterGroupIDWhenAbsent(t *testing.T) {
	r := &CacheResource{}
	data := &CacheResourceModel{
		ParameterGroupID: types.StringValue("pg-123"),
	}
	cache := &client.CacheInstance{
		ID:               "cache-1",
		Name:             "my-cache",
		Status:           "running",
		Provider:         client.CacheProvider{ID: 1, Name: "Redis", Type: "redis"},
		ParameterGroupID: nil,
	}

	r.mapCacheToState(cache, data)

	if data.ParameterGroupID.IsNull() {
		t.Fatal("ParameterGroupID = null, want preserved value pg-123")
	}
	if data.ParameterGroupID.ValueString() != "pg-123" {
		t.Errorf("ParameterGroupID = %v, want pg-123", data.ParameterGroupID.ValueString())
	}
}

func TestMapCacheToState_UpdatesParameterGroupIDWhenPresent(t *testing.T) {
	r := &CacheResource{}
	data := &CacheResourceModel{
		ParameterGroupID: types.StringValue("pg-old"),
	}
	newID := "pg-new"
	cache := &client.CacheInstance{
		ID:               "cache-1",
		Name:             "my-cache",
		Status:           "running",
		Provider:         client.CacheProvider{ID: 1, Name: "Redis", Type: "redis"},
		ParameterGroupID: &newID,
	}

	r.mapCacheToState(cache, data)

	if data.ParameterGroupID.ValueString() != "pg-new" {
		t.Errorf("ParameterGroupID = %v, want pg-new", data.ParameterGroupID.ValueString())
	}
}

func TestMapCacheToState_ProviderPrefersType(t *testing.T) {
	r := &CacheResource{}
	data := &CacheResourceModel{}
	cache := &client.CacheInstance{
		ID:       "cache-1",
		Name:     "my-cache",
		Status:   "running",
		Provider: client.CacheProvider{ID: 1, Name: "Redis", Type: "redis"},
	}

	r.mapCacheToState(cache, data)

	if data.CacheProvider.ValueString() != "redis" {
		t.Errorf("CacheProvider = %v, want redis", data.CacheProvider.ValueString())
	}
}

func TestMapCacheToState_ProviderFallsBackToLowercasedName(t *testing.T) {
	r := &CacheResource{}
	data := &CacheResourceModel{}
	cache := &client.CacheInstance{
		ID:       "cache-1",
		Name:     "my-cache",
		Status:   "running",
		Provider: client.CacheProvider{ID: 1, Name: "Redis"},
	}

	r.mapCacheToState(cache, data)

	if data.CacheProvider.ValueString() != "redis" {
		t.Errorf("CacheProvider = %v, want redis (lowercased fallback)", data.CacheProvider.ValueString())
	}
}
