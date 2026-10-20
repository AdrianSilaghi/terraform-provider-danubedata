package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestVpsCreateRequiresExplicitCPUPlatform(t *testing.T) {
	var schemaResp resource.SchemaResponse
	(&VpsResource{}).Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	platform := schemaResp.Schema.Attributes["cpu_platform"]
	if platform == nil || !platform.IsRequired() {
		t.Fatal("cpu_platform must be required because the API has no default")
	}
	instance := &client.VpsInstance{ID: "vps-1", CPUPlatform: "intel"}
	var state VpsResourceModel
	(&VpsResource{}).mapVpsToState(instance, &state)
	if state.CPUPlatform.ValueString() != "intel" {
		t.Fatalf("cpu_platform = %q", state.CPUPlatform.ValueString())
	}
}

func TestFirewallRuleRoundTripPreservesSinglePort(t *testing.T) {
	port := 22
	fw := &client.Firewall{ID: "fw-1", Name: "fw", Rules: []client.FirewallRule{{ID: "r-1", Name: "SSH", Action: "allow", Direction: "inbound", Protocol: "tcp", Port: &port, IsLocked: false}}}
	var state FirewallResourceModel
	var diags diag.Diagnostics
	(&FirewallResource{}).mapFirewallToState(context.Background(), fw, &state, &diags)
	if diags.HasError() {
		t.Fatalf("mapping failed: %v", diags)
	}
	var rules []FirewallRuleModel
	diags.Append(state.Rules.ElementsAs(context.Background(), &rules, false)...)
	if diags.HasError() || len(rules) != 1 {
		t.Fatalf("rules: %v %v", rules, diags)
	}
	if !rules[0].Port.Equal(types.Int64Value(22)) || !rules[0].IsLocked.Equal(types.BoolValue(false)) {
		t.Fatalf("single port or locked flag lost: %#v", rules[0])
	}
}

func TestVpsImageChangeAppearsAsDrift(t *testing.T) {
	state := VpsResourceModel{Image: types.StringValue("ubuntu-24.04")}
	(&VpsResource{}).mapVpsToState(&client.VpsInstance{Image: "registry.example/kubevirt-debian:debian-13-2026.10.01"}, &state)
	if got := state.Image.ValueString(); got != "debian-13" {
		t.Fatalf("image drift hidden: %q", got)
	}
}

func TestFirewallUnresolvedSourceCannotBeWritten(t *testing.T) {
	fw := &client.Firewall{ID: "fw-1", Rules: []client.FirewallRule{{ID: "r-1", Action: "allow", Direction: "inbound", Protocol: "tcp", SourceInstanceIDs: []client.FirewallSourceInstance{{ID: "orphan"}}}}}
	var state FirewallResourceModel
	var diags diag.Diagnostics
	(&FirewallResource{}).mapFirewallToState(context.Background(), fw, &state, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	_, diags = firewallRulesToRequest(context.Background(), state.Rules)
	if !diags.HasError() {
		t.Fatal("unresolved source must fail closed")
	}
}

func TestLockedFirewallRulesStayOutsideManagedRuleList(t *testing.T) {
	port := 25
	fw := &client.Firewall{ID: "fw-1", Rules: []client.FirewallRule{{ID: "locked", Action: "deny", Direction: "outbound", Protocol: "tcp", Port: &port, IsLocked: true}}}
	var state FirewallResourceModel
	var diags diag.Diagnostics
	(&FirewallResource{}).mapFirewallToState(context.Background(), fw, &state, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if len(state.Rules.Elements()) != 0 {
		t.Fatalf("locked rule entered managed list: %v", state.Rules)
	}
}

func TestVpsCreateRetainsIDWhenWaitFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/vps" {
			_, _ = w.Write([]byte(`{"instance":{"id":"paid-vps"}}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	ctx := context.Background()
	target := &VpsResource{client: client.New(client.Config{BaseURL: server.URL, APIToken: "test"})}
	var schemaResp resource.SchemaResponse
	target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	model := VpsResourceModel{Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType})}, Name: types.StringValue("unit-vps"), ResourceProfile: types.StringValue("nano_shared"), CPUPlatform: types.StringValue("amd"), Image: types.StringValue("ubuntu-24.04"), Datacenter: types.StringValue("fsn1"), AuthMethod: types.StringValue("password"), Password: types.StringValue("Testing123!abc")}
	if d := plan.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	if d := resp.State.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	target.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("wait failure must be reported")
	}
	var id types.String
	if d := resp.State.GetAttribute(ctx, path.Root("id"), &id); d.HasError() {
		t.Fatal(d)
	}
	if id.ValueString() != "paid-vps" {
		t.Fatalf("paid VPS ID lost after waiter failure: %q", id.ValueString())
	}
}

func TestVpsSnapshotUnknownSizeIsNull(t *testing.T) {
	var state VpsSnapshotResourceModel
	(&VpsSnapshotResource{}).mapSnapshotToState(&client.VpsSnapshot{ID: 42, Status: "creating"}, &state)
	if !state.SizeGB.IsNull() {
		t.Fatalf("unknown size reported as %v GB", state.SizeGB)
	}
}

func TestFirewallReorderedRuleCannotInheritAnotherRulesSources(t *testing.T) {
	ctx := context.Background()
	fw := &client.Firewall{Rules: []client.FirewallRule{
		{ID: "ssh", Name: "SSH", Action: "allow", Direction: "inbound", Protocol: "tcp", SourceInstanceIDs: []client.FirewallSourceInstance{{ID: "trusted", Type: stringPointer("vps")}}},
		{ID: "http", Name: "HTTP", Action: "allow", Direction: "inbound", Protocol: "tcp", SourceInstanceIDs: []client.FirewallSourceInstance{{ID: "public", Type: stringPointer("vps")}}},
	}}
	var state FirewallResourceModel
	var diags diag.Diagnostics
	(&FirewallResource{}).mapFirewallToState(ctx, fw, &state, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	var planned []FirewallRuleModel
	diags.Append(state.Rules.ElementsAs(ctx, &planned, false)...)
	if diags.HasError() {
		t.Fatal(diags)
	}
	planned[0], planned[1] = planned[1], planned[0]
	for i := range planned {
		planned[i].ID = types.StringUnknown()
		planned[i].SourceInstanceIDs = types.ListUnknown(types.ObjectType{AttrTypes: firewallSourceTypes})
	}
	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: firewallRuleTypes}, planned)
	diags.Append(d...)
	if diags.HasError() {
		t.Fatal(diags)
	}
	rules, d := firewallRulesToRequest(ctx, list, state.Rules)
	if d.HasError() {
		t.Fatal(d)
	}
	if len(rules) != 2 || rules[0].SourceInstanceIDs[0].ID != "public" || rules[1].SourceInstanceIDs[0].ID != "trusted" {
		t.Fatalf("reordered rules inherited the wrong source restrictions: %#v", rules)
	}
	planned[0].ID = types.StringValue("ssh")
	planned[1].ID = types.StringValue("http")
	list, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: firewallRuleTypes}, planned)
	if d.HasError() {
		t.Fatal(d)
	}
	_, d = firewallRulesToRequest(ctx, list, state.Rules)
	if !d.HasError() {
		t.Fatal("positional computed IDs must not silently change a reordered rule's restrictions")
	}
}

func stringPointer(value string) *string { return &value }
