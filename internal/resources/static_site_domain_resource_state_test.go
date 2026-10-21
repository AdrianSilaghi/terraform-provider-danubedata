package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestStaticSiteDomainStateUpgradePreservesAttachment(t *testing.T) {
	ctx := context.Background()
	r := &StaticSiteDomainResource{}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	if schema.Schema.Version != 1 {
		t.Fatalf("schema version = %d", schema.Schema.Version)
	}

	prior := []byte(`{"id":"site-1:domain-1","static_site_id":"site-1","domain_id":"domain-1","domain":"www.example.com","verification_status":"pending","tls_status":"pending","deployment_status":"pending","is_primary":false,"dns_instructions":{"record_type":"TXT","record_name":"_verify","record_value":"abc","instructions":"Add TXT"},"created_at":"2026-01-01"}`)
	resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.UpgradeState(ctx)[0].StateUpgrader(ctx, resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: prior}}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	var got StaticSiteDomainResourceModel
	if diags := resp.State.Get(ctx, &got); diags.HasError() {
		t.Fatal(diags)
	}
	if got.ID.ValueString() != "site-1:domain-1" || got.Domain.ValueString() != "www.example.com" {
		t.Fatalf("attachment changed: id=%q domain=%q", got.ID.ValueString(), got.Domain.ValueString())
	}
	if !got.DNSInstructions.IsNull() {
		t.Fatalf("old computed DNS object survived: %v", got.DNSInstructions)
	}
}

func TestStaticSiteDomainStateUpgradeRejectsUnusableState(t *testing.T) {
	ctx := context.Background()
	r := &StaticSiteDomainResource{}
	var current resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &current)
	for _, tc := range []struct {
		name string
		raw  *tfprotov6.RawState
	}{
		{name: "missing", raw: nil},
		{name: "malformed", raw: &tfprotov6.RawState{JSON: []byte(`{"domain":`)}},
		{name: "wrong shape", raw: &tfprotov6.RawState{JSON: []byte(`[]`)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: current.Schema}}
			r.UpgradeState(ctx)[0].StateUpgrader(ctx, resource.UpgradeStateRequest{RawState: tc.raw}, &resp)
			if !resp.Diagnostics.HasError() {
				t.Fatal("unusable prior state must fail instead of losing the attachment")
			}
		})
	}
}
