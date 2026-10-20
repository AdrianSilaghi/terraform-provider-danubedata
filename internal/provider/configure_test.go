package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestConfigureProjectScope(t *testing.T) {
	for _, tc := range []struct {
		name string
		team interface{}
		env  string
		want string
		fail bool
	}{
		{name: "token default"},
		{name: "environment", env: "42", want: "42"},
		{name: "configuration overrides environment", team: int64(7), env: "invalid", want: "7"},
		{name: "invalid environment", env: "invalid", fail: true},
		{name: "zero environment", env: "0", fail: true},
		{name: "negative configuration", team: int64(-1), fail: true},
		{name: "unknown configuration", team: tftypes.UnknownValue, env: "42", fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DANUBEDATA_TEAM_ID", tc.env)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("X-Team-Id"); got != tc.want {
					t.Errorf("project header = %q, want %q", got, tc.want)
				}
				fmt.Fprint(w, `{"data":[],"pagination":{"last_page":1}}`)
			}))
			defer server.Close()
			p := New("test")()
			var schema provider.SchemaResponse
			p.Schema(context.Background(), provider.SchemaRequest{}, &schema)
			raw := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"base_url": tftypes.String, "api_token": tftypes.String, "team_id": tftypes.Number,
			}}, map[string]tftypes.Value{
				"base_url":  tftypes.NewValue(tftypes.String, server.URL),
				"api_token": tftypes.NewValue(tftypes.String, "fake-token"),
				"team_id":   tftypes.NewValue(tftypes.Number, tc.team),
			})
			var resp provider.ConfigureResponse
			p.Configure(context.Background(), provider.ConfigureRequest{Config: tfsdk.Config{Raw: raw, Schema: schema.Schema}}, &resp)
			if resp.Diagnostics.HasError() != tc.fail {
				t.Fatalf("diagnostics = %v, want error %v", resp.Diagnostics, tc.fail)
			}
			if !tc.fail {
				c, ok := resp.ResourceData.(*client.Client)
				if !ok {
					t.Fatalf("client = %T", resp.ResourceData)
				}
				if _, err := c.ListSshKeys(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
