package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestFirewallUnknownPortCannotWidenRule(t *testing.T) {
	ctx := context.Background()
	port := 22
	start, end := 80, 443
	for _, tc := range []struct {
		name   string
		change func(*FirewallRuleModel)
	}{
		{"single port", func(rule *FirewallRuleModel) { rule.Port = types.Int64Unknown() }},
		{"range start", func(rule *FirewallRuleModel) { rule.PortRangeStart = types.Int64Unknown() }},
		{"range end", func(rule *FirewallRuleModel) { rule.PortRangeEnd = types.Int64Unknown() }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rule := client.FirewallRule{ID: "rule-id", Name: "SSH", Action: "allow", Direction: "inbound", Protocol: "tcp", Port: &port}
			if tc.name != "single port" {
				rule.Port, rule.PortRangeStart, rule.PortRangeEnd = nil, &start, &end
			}
			fw := &client.Firewall{Rules: []client.FirewallRule{rule}}
			var state FirewallResourceModel
			var diags diag.Diagnostics
			(&FirewallResource{}).mapFirewallToState(ctx, fw, &state, &diags)
			if diags.HasError() {
				t.Fatal(diags)
			}
			var rules []FirewallRuleModel
			if d := state.Rules.ElementsAs(ctx, &rules, false); d.HasError() {
				t.Fatal(d)
			}
			tc.change(&rules[0])
			planned, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: firewallRuleTypes}, rules)
			if d.HasError() {
				t.Fatal(d)
			}
			if _, d := firewallRulesToRequest(ctx, planned, state.Rules); !d.HasError() {
				t.Fatal("unresolved port must not be sent as an unrestricted rule")
			}
		})
	}
}

func TestFirewallUnknownSourceIDCannotBeWritten(t *testing.T) {
	ctx := context.Background()
	fw := &client.Firewall{Rules: []client.FirewallRule{{ID: "rule-id", Name: "restricted", Action: "allow", Direction: "inbound", Protocol: "tcp", SourceInstanceIDs: []client.FirewallSourceInstance{{ID: "source-id", Type: stringPointer("vps")}}}}}
	var state FirewallResourceModel
	var diags diag.Diagnostics
	(&FirewallResource{}).mapFirewallToState(ctx, fw, &state, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	var rules []FirewallRuleModel
	if d := state.Rules.ElementsAs(ctx, &rules, false); d.HasError() {
		t.Fatal(d)
	}
	sources, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: firewallSourceTypes}, []struct {
		ID   types.String `tfsdk:"id"`
		Type types.String `tfsdk:"type"`
	}{{ID: types.StringUnknown(), Type: types.StringValue("vps")}})
	if d.HasError() {
		t.Fatal(d)
	}
	rules[0].SourceInstanceIDs = sources
	planned, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: firewallRuleTypes}, rules)
	if d.HasError() {
		t.Fatal(d)
	}
	if _, d := firewallRulesToRequest(ctx, planned, state.Rules); !d.HasError() {
		t.Fatal("unknown source ID must not be sent as an empty source")
	}
}

func TestVpsSnapshotDeleteHandlesInvalidAndMissingIDs(t *testing.T) {
	ctx := context.Background()
	called := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		if r.Method != http.MethodDelete || r.URL.Path != "/snapshots/vps/42" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	target := &VpsSnapshotResource{client: client.New(client.Config{BaseURL: server.URL, APIToken: "test"})}
	var schemaResp resource.SchemaResponse
	target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	for _, tc := range []struct {
		id, wantError string
		wantRequests  int
	}{
		{"invalid", "Invalid VPS snapshot ID", 0},
		{"42", "", 1},
	} {
		t.Run(tc.id, func(t *testing.T) {
			state := tfsdk.State{Schema: schemaResp.Schema}
			model := VpsSnapshotResourceModel{ID: types.StringValue(tc.id), Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "delete": types.StringType})}}
			if d := state.Set(ctx, &model); d.HasError() {
				t.Fatal(d)
			}
			resp := &resource.DeleteResponse{}
			target.Delete(ctx, resource.DeleteRequest{State: state}, resp)
			if (tc.wantError != "") != resp.Diagnostics.HasError() {
				t.Fatalf("diagnostics = %v", resp.Diagnostics)
			}
			if called != tc.wantRequests {
				t.Fatalf("requests = %d, want %d", called, tc.wantRequests)
			}
		})
	}
}

func TestVpsSnapshotMappingClearsExpiredValues(t *testing.T) {
	expires := "2026-12-31"
	state := VpsSnapshotResourceModel{ExpiresAt: types.StringValue(expires), SizeGB: types.Float64Value(1.5), SizeBytes: types.Int64Value(1500000000), SizeBasis: types.StringValue("stored")}
	(&VpsSnapshotResource{}).mapSnapshotToState(&client.VpsSnapshot{ID: 42, Name: "backup", Status: "creating"}, &state)
	if !state.ExpiresAt.IsNull() || !state.SizeGB.IsNull() || !state.SizeBytes.IsNull() || !state.SizeBasis.IsNull() {
		t.Fatalf("stale snapshot values remained: %#v", state)
	}
}

func TestVpsPasswordKeepsConfiguredSecret(t *testing.T) {
	called := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		if r.Method != http.MethodGet || r.URL.Path != "/vps/vps-1/password" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"password":"api-generated"}`))
	}))
	defer server.Close()
	target := &VpsResource{client: client.New(client.Config{BaseURL: server.URL, APIToken: "test"})}
	configured := VpsResourceModel{Password: types.StringValue("configured-secret")}
	target.fetchVpsPassword(context.Background(), "vps-1", &configured)
	if configured.Password.ValueString() != "configured-secret" || called != 0 {
		t.Fatalf("configured password was replaced or fetched: %q, requests=%d", configured.Password.ValueString(), called)
	}
	computed := VpsResourceModel{Password: types.StringNull()}
	target.fetchVpsPassword(context.Background(), "vps-1", &computed)
	if computed.Password.ValueString() != "api-generated" || called != 1 {
		t.Fatalf("computed password was not read: %q, requests=%d", computed.Password.ValueString(), called)
	}
}

func TestVpsCreateSendsConfiguredNetworkAndCloudInit(t *testing.T) {
	ctx := context.Background()
	created := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /vps":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			for key, want := range map[string]any{
				"network_tier": "10g", "network_stack": "ipv6_only", "cpu_allocation_type": "dedicated",
				"primary_ipv4_floating_ip_id": "floating-v4", "primary_ipv6_floating_ip_id": "floating-v6",
				"custom_cloud_init": "#cloud-config", "password": "configured-secret", "password_confirmation": "configured-secret",
			} {
				if body[key] != want {
					t.Errorf("create %s = %v, want %v", key, body[key], want)
				}
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"instance":{"id":"vps-1"}}`))
		case "GET /vps/vps-1/status":
			_, _ = w.Write([]byte(`{"status":"running"}`))
		case "GET /vps/vps-1":
			_, _ = w.Write([]byte(`{"instance":{"id":"vps-1","name":"unit-vps","status":"running","resource_profile":"nano_dedicated","cpu_platform":"amd","network_tier":"10g","cpu_allocation_type":"dedicated","image":"ubuntu-24.04","datacenter":"fsn1"}}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	target := &VpsResource{client: client.New(client.Config{BaseURL: server.URL, APIToken: "test"})}
	var schemaResp resource.SchemaResponse
	target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	model := VpsResourceModel{
		Name: types.StringValue("unit-vps"), ResourceProfile: types.StringValue("nano_dedicated"), Image: types.StringValue("ubuntu-24.04"), Datacenter: types.StringValue("fsn1"),
		CPUPlatform: types.StringValue("amd"), CPUAllocationType: types.StringValue("dedicated"), NetworkTier: types.StringValue("10g"), NetworkStack: types.StringValue("ipv6_only"),
		PrimaryIPv4FloatingIPID: types.StringValue("floating-v4"), PrimaryIPv6FloatingIPID: types.StringValue("floating-v6"),
		AuthMethod: types.StringValue("password"), Password: types.StringValue("configured-secret"), CustomCloudInit: types.StringValue("#cloud-config"),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType})},
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if d := plan.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	if d := resp.State.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	target.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	if resp.Diagnostics.HasError() || !created {
		t.Fatalf("create failed: %v", resp.Diagnostics)
	}
	var result VpsResourceModel
	if d := resp.State.Get(ctx, &result); d.HasError() {
		t.Fatal(d)
	}
	if result.Password.ValueString() != "configured-secret" || result.NetworkStack.ValueString() != "ipv6_only" {
		t.Fatalf("write-only fields were not preserved: password=%q, network_stack=%q", result.Password.ValueString(), result.NetworkStack.ValueString())
	}
}

func TestVpsSnapshotTimeoutUpdatePreservesComputedState(t *testing.T) {
	ctx := context.Background()
	target := &VpsSnapshotResource{}
	var schemaResp resource.SchemaResponse
	target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	model := VpsSnapshotResourceModel{
		ID: types.StringValue("42"), Name: types.StringValue("backup"), VpsInstanceID: types.StringValue("vps-1"),
		Status: types.StringValue("ready"), SizeGB: types.Float64Value(1.25), SizeBytes: types.Int64Value(1342177280),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "delete": types.StringType})},
	}
	state := tfsdk.State{Schema: schemaResp.Schema}
	if d := state.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: state.Raw}
	newTimeouts := timeouts.Value{Object: types.ObjectValueMust(map[string]attr.Type{"create": types.StringType, "delete": types.StringType}, map[string]attr.Value{"create": types.StringValue("45m"), "delete": types.StringValue("5m")})}
	if d := plan.SetAttribute(ctx, path.Root("timeouts"), newTimeouts); d.HasError() {
		t.Fatal(d)
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	target.Update(ctx, resource.UpdateRequest{State: state, Plan: plan}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	var result VpsSnapshotResourceModel
	if d := resp.State.Get(ctx, &result); d.HasError() {
		t.Fatal(d)
	}
	if result.ID.ValueString() != "42" || result.Status.ValueString() != "ready" || result.SizeBytes.ValueInt64() != 1342177280 || result.Timeouts.IsNull() {
		t.Fatalf("timeout update lost snapshot state: %#v", result)
	}
}

func TestVpsDeleteStopsBeforeDestroying(t *testing.T) {
	ctx := context.Background()
	stops, deletes := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /vps/vps-1/status":
			_, _ = w.Write([]byte(`{"status":"running"}`))
		case "POST /vps/vps-1/stop":
			stops++
			w.WriteHeader(http.StatusInternalServerError)
		case "DELETE /vps/vps-1":
			deletes++
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	target := &VpsResource{client: client.New(client.Config{BaseURL: server.URL, APIToken: "test"})}
	var schemaResp resource.SchemaResponse
	target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}
	model := VpsResourceModel{ID: types.StringValue("vps-1"), Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType})}}
	if d := state.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	resp := &resource.DeleteResponse{}
	target.Delete(ctx, resource.DeleteRequest{State: state}, resp)
	if !resp.Diagnostics.HasError() || stops != 1 || deletes != 0 {
		t.Fatalf("unsafe delete after stop failure: diagnostics=%v, stops=%d, deletes=%d", resp.Diagnostics, stops, deletes)
	}
}

func TestFirewallDeployFailureRetainsPriorState(t *testing.T) {
	ctx := context.Background()
	puts, deploys := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "PUT /firewalls/fw-1":
			puts++
			_, _ = w.Write([]byte(`{"firewall":{"id":"fw-1","name":"firewall","rules":[{"id":"rule-id","name":"SSH","action":"allow","direction":"inbound","protocol":"tcp","port":23}]}}`))
		case "POST /firewalls/fw-1/deploy":
			deploys++
			w.WriteHeader(http.StatusInternalServerError)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	target := &FirewallResource{client: client.New(client.Config{BaseURL: server.URL, APIToken: "test"})}
	var schemaResp resource.SchemaResponse
	target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	port := 22
	fw := &client.Firewall{ID: "fw-1", Name: "firewall", Rules: []client.FirewallRule{{ID: "rule-id", Name: "SSH", Action: "allow", Direction: "inbound", Protocol: "tcp", Port: &port}}}
	model := FirewallResourceModel{}
	var diags diag.Diagnostics
	target.mapFirewallToState(ctx, fw, &model, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	state := tfsdk.State{Schema: schemaResp.Schema}
	if d := state.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	var rules []FirewallRuleModel
	if d := model.Rules.ElementsAs(ctx, &rules, false); d.HasError() {
		t.Fatal(d)
	}
	rules[0].Port = types.Int64Value(23)
	planRules, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: firewallRuleTypes}, rules)
	if d.HasError() {
		t.Fatal(d)
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: state.Raw}
	if d := plan.SetAttribute(ctx, path.Root("rules"), planRules); d.HasError() {
		t.Fatal(d)
	}
	resp := &resource.UpdateResponse{State: state}
	target.Update(ctx, resource.UpdateRequest{State: state, Plan: plan}, resp)
	if !resp.Diagnostics.HasError() || puts != 1 || deploys != 1 {
		t.Fatalf("deploy failure was hidden: diagnostics=%v, puts=%d, deploys=%d", resp.Diagnostics, puts, deploys)
	}
	if !strings.Contains(resp.Diagnostics.Errors()[0].Detail(), "POST /api/v1/firewalls/fw-1/deploy") {
		t.Fatalf("deploy failure gave no recovery action: %v", resp.Diagnostics)
	}
	var result FirewallResourceModel
	if d := resp.State.Get(ctx, &result); d.HasError() {
		t.Fatal(d)
	}
	var oldRules []FirewallRuleModel
	if d := result.Rules.ElementsAs(ctx, &oldRules, false); d.HasError() || len(oldRules) != 1 || oldRules[0].Port.ValueInt64() != 22 {
		t.Fatalf("failed deployment changed tracked rule state: %#v, %v", oldRules, d)
	}
}

func TestVpsUpdateWarnsWhenResizeIsOnlyQueued(t *testing.T) {
	ctx := context.Background()
	puts, reads := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "PUT /vps/vps-1":
			puts++
			_, _ = w.Write([]byte(`{"instance":{"id":"vps-1"}}`))
		case "GET /vps/vps-1":
			reads++
			_, _ = w.Write([]byte(`{"instance":{"id":"vps-1","name":"unit-vps","status":"stopped","resource_profile":"micro_shared","cpu_platform":"amd","image":"ubuntu-24.04","datacenter":"fsn1"}}`))
		case "GET /vps/vps-1/password":
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	target := &VpsResource{client: client.New(client.Config{BaseURL: server.URL, APIToken: "test"})}
	var schemaResp resource.SchemaResponse
	target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	model := VpsResourceModel{
		ID: types.StringValue("vps-1"), Name: types.StringValue("unit-vps"), Status: types.StringValue("stopped"),
		ResourceProfile: types.StringValue("nano_shared"), CPUPlatform: types.StringValue("amd"), Image: types.StringValue("ubuntu-24.04"), Datacenter: types.StringValue("fsn1"),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType})},
	}
	state := tfsdk.State{Schema: schemaResp.Schema}
	if d := state.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: state.Raw}
	if d := plan.SetAttribute(ctx, path.Root("resource_profile"), "micro_shared"); d.HasError() {
		t.Fatal(d)
	}
	resp := &resource.UpdateResponse{State: state}
	target.Update(ctx, resource.UpdateRequest{State: state, Plan: plan}, resp)
	if resp.Diagnostics.HasError() || puts != 1 || reads != 1 {
		t.Fatalf("queued resize should refresh once: diagnostics=%v, puts=%d, reads=%d", resp.Diagnostics, puts, reads)
	}
	if len(resp.Diagnostics.Warnings()) == 0 || !strings.Contains(resp.Diagnostics.Warnings()[0].Detail(), "queued") {
		t.Fatalf("queued resize lacks warning: %v", resp.Diagnostics)
	}
}
