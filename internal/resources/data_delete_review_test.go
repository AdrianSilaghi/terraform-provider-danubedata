package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDataDeleteReviewRetainsStateUntilDeletionConfirmed(t *testing.T) {
	type reply struct {
		method, suffix, instanceStatus string
		status                         int
	}
	get := func(status int, instanceStatus string) reply {
		return reply{method: http.MethodGet, status: status, instanceStatus: instanceStatus}
	}
	stop := func(status int) reply { return reply{method: http.MethodPost, suffix: "/stop", status: status} }
	remove := func(status int) reply { return reply{method: http.MethodDelete, status: status} }
	cases := []struct {
		name    string
		steps   []reply
		wantErr bool
		timeout bool
	}{
		{"read denied", []reply{get(http.StatusForbidden, "")}, true, false},
		{"read failed", []reply{get(http.StatusInternalServerError, "")}, true, false},
		{"already gone", []reply{get(http.StatusNotFound, "")}, false, false},
		{"stop denied", []reply{get(http.StatusOK, "running"), stop(http.StatusForbidden)}, true, false},
		{"stop failed", []reply{get(http.StatusOK, "running"), stop(http.StatusInternalServerError)}, true, false},
		{"stop wait canceled", []reply{get(http.StatusOK, "running"), stop(http.StatusOK), get(http.StatusOK, "running")}, true, true},
		{"stop status read failed", []reply{get(http.StatusOK, "running"), stop(http.StatusOK), get(http.StatusInternalServerError, "")}, true, false},
		{"delete denied", []reply{get(http.StatusOK, "stopped"), remove(http.StatusForbidden)}, true, false},
		{"delete failed", []reply{get(http.StatusOK, "stopped"), remove(http.StatusInternalServerError)}, true, false},
		{"delete says missing", []reply{get(http.StatusOK, "stopped"), remove(http.StatusNotFound)}, false, false},
		{"delete confirmed", []reply{get(http.StatusOK, "stopped"), remove(http.StatusOK), get(http.StatusNotFound, "")}, false, false},
		{"delete wait canceled", []reply{get(http.StatusOK, "stopped"), remove(http.StatusOK), get(http.StatusOK, "stopped")}, true, true},
		{"delete status read failed", []reply{get(http.StatusOK, "stopped"), remove(http.StatusOK), get(http.StatusInternalServerError, "")}, true, false},
		{"running stops before delete", []reply{get(http.StatusOK, "running"), stop(http.StatusOK), get(http.StatusOK, "stopped"), remove(http.StatusOK), get(http.StatusNotFound, "")}, false, false},
		{"pending becomes error", []reply{get(http.StatusOK, "pending"), get(http.StatusOK, "error"), get(http.StatusOK, "error"), remove(http.StatusOK), get(http.StatusNotFound, "")}, false, false},
		{"pending disappears", []reply{get(http.StatusOK, "pending"), get(http.StatusNotFound, ""), get(http.StatusNotFound, "")}, false, false},
		{"pending remains pending", []reply{get(http.StatusOK, "pending"), get(http.StatusOK, "pending")}, true, true},
	}
	for _, kind := range []string{"cache", "database"} {
		for _, tc := range cases {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				id := kind + "-1"
				calls := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if calls >= len(tc.steps) {
						t.Errorf("unexpected extra request %s %s", r.Method, r.URL.Path)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					step := tc.steps[calls]
					calls++
					if r.Method != step.method || r.URL.Path != "/"+kind+"/"+id+step.suffix {
						t.Errorf("request %d = %s %s, want %s /%s/%s%s", calls, r.Method, r.URL.Path, step.method, kind, id, step.suffix)
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(step.status)
					if step.status == http.StatusOK && step.method == http.MethodGet {
						_ = json.NewEncoder(w).Encode(map[string]any{"instance": map[string]string{"id": id, "status": step.instanceStatus}})
					}
				}))
				defer server.Close()

				ctx := context.Background()
				if tc.timeout {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, 200*time.Millisecond)
					defer cancel()
				}
				c := client.New(client.Config{BaseURL: server.URL, APIToken: "offline-token"})
				var schemaResp resource.SchemaResponse
				var deleteResource func(context.Context, resource.DeleteRequest, *resource.DeleteResponse)
				timeoutValue := timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType})}
				state := tfsdk.State{}
				if kind == "cache" {
					target := &CacheResource{client: c}
					target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
					state.Schema = schemaResp.Schema
					if diags := state.Set(ctx, &CacheResourceModel{ID: types.StringValue(id), Configuration: types.MapNull(types.StringType), Timeouts: timeoutValue}); diags.HasError() {
						t.Fatal(diags)
					}
					deleteResource = target.Delete
				} else {
					target := &DatabaseResource{client: c}
					target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
					state.Schema = schemaResp.Schema
					if diags := state.Set(ctx, &DatabaseResourceModel{ID: types.StringValue(id), Configuration: types.MapNull(types.StringType), Timeouts: timeoutValue}); diags.HasError() {
						t.Fatal(diags)
					}
					deleteResource = target.Delete
				}
				resp := &resource.DeleteResponse{State: state}
				deleteResource(ctx, resource.DeleteRequest{State: state}, resp)
				if resp.Diagnostics.HasError() != tc.wantErr {
					t.Fatalf("diagnostics = %v, want error %v", resp.Diagnostics, tc.wantErr)
				}
				if calls != len(tc.steps) {
					t.Fatalf("requests = %d, want %d", calls, len(tc.steps))
				}
				var trackedID types.String
				if diags := resp.State.GetAttribute(ctx, path.Root("id"), &trackedID); diags.HasError() || trackedID.ValueString() != id {
					t.Fatalf("tracked ID changed during delete: %q, %v", trackedID.ValueString(), diags)
				}
			})
		}
	}
}
