package datasources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestFirewallsDataSourceCountsRulesWhenListOmitsThem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/firewalls":
			_, _ = w.Write([]byte(`{"data":[{"id":"fw-1","name":"web","is_system":true}],"pagination":{"last_page":1}}`))
		case "/firewalls/fw-1":
			_, _ = w.Write([]byte(`{"firewall":{"id":"fw-1","name":"web","rules":[{"id":"r-1","action":"allow","direction":"inbound","protocol":"tcp"}]}}`))
		default:
			t.Errorf("unexpected %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	ctx := context.Background()
	d := &FirewallsDataSource{client: client.New(client.Config{BaseURL: server.URL})}
	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	d.Read(ctx, datasource.ReadRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	var data FirewallsDataSourceModel
	if diags := resp.State.Get(ctx, &data); diags.HasError() {
		t.Fatal(diags)
	}
	if len(data.Firewalls) != 1 || data.Firewalls[0].RulesCount.ValueInt64() != 1 || !data.Firewalls[0].IsSystem.ValueBool() {
		t.Fatalf("firewalls = %#v", data.Firewalls)
	}
}
