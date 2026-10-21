package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestImportedPrivateGitRuntimeEditPreservesAuth(t *testing.T) {
	ctx := context.Background()
	name, generation, updates := "original", 1, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /serverless/private-1":
			_ = json.NewEncoder(w).Encode(map[string]any{"container": map[string]any{
				"id": "private-1", "name": name, "slug": "private", "status": "running", "resource_profile": "small",
				"deployment_type": "git_repository", "source_type": "dockerfile", "repository_url": "https://example.com/org/app",
				"repository_branch": "main", "git_auth_type": "access_token", "image_tag": "latest",
				"port": 8080, "min_scale": 0, "max_scale": 10, "environment_variables": map[string]string{},
				"spec_generation": generation, "observed_generation": generation,
				"status_details": map[string]any{"summary": "ready", "operation": map[string]bool{"terminal": true}},
			}, "url": "https://private.example", "monthly_cost": "0.00"})
		case "PUT /serverless/private-1":
			updates++
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "renamed" {
				t.Errorf("name = %v", body["name"])
			}
			if _, ok := body["git_auth_type"]; ok {
				t.Errorf("unrelated edit changed imported Git auth: %v", body)
			}
			name, generation = "renamed", 2
			_, _ = fmt.Fprint(w, `{"container":{"id":"private-1","spec_generation":2}}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	target := &ServerlessResource{client: client.New(client.Config{BaseURL: server.URL, APIToken: "test"})}
	var schemaResp resource.SchemaResponse
	target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	auth := schemaResp.Schema.Attributes["git_auth_type"].(schema.StringAttribute)
	if auth.Default != nil || len(auth.PlanModifiers) == 0 {
		t.Fatal("omitted Git auth must keep imported state during planning")
	}
	prior := tfsdk.State{Schema: schemaResp.Schema}
	if diags := prior.Set(ctx, &ServerlessResourceModel{
		ID: types.StringValue("private-1"), EnvironmentVariables: types.MapNull(types.StringType),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType})},
	}); diags.HasError() {
		t.Fatal(diags)
	}
	read := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	target.Read(ctx, resource.ReadRequest{State: prior}, read)
	if read.Diagnostics.HasError() {
		t.Fatal(read.Diagnostics)
	}
	var imported ServerlessResourceModel
	if diags := read.State.Get(ctx, &imported); diags.HasError() {
		t.Fatal(diags)
	}
	if imported.GitAuthType.ValueString() != "access_token" {
		t.Fatalf("imported auth = %q", imported.GitAuthType.ValueString())
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: read.State.Raw}
	if diags := plan.SetAttribute(ctx, path.Root("name"), "renamed"); diags.HasError() {
		t.Fatal(diags)
	}
	updated := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	target.Update(ctx, resource.UpdateRequest{State: read.State, Plan: plan}, updated)
	if updated.Diagnostics.HasError() {
		t.Fatal(updated.Diagnostics)
	}
	var got ServerlessResourceModel
	if diags := updated.State.Get(ctx, &got); diags.HasError() {
		t.Fatal(diags)
	}
	if updates != 1 || got.Name.ValueString() != "renamed" || got.GitAuthType.ValueString() != "access_token" {
		t.Fatalf("runtime edit lost imported auth: updates=%d name=%q auth=%q", updates, got.Name.ValueString(), got.GitAuthType.ValueString())
	}
}
