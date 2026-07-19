package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMapDatabaseToState_PreservesParameterGroupIDWhenAbsent(t *testing.T) {
	r := &DatabaseResource{}
	data := &DatabaseResourceModel{
		ParameterGroupID: types.StringValue("pg-123"),
	}
	database := &client.DatabaseInstance{
		ID:               "db-1",
		Name:             "my-database",
		Status:           "running",
		Engine:           client.DatabaseEngine{ID: 1, Name: "mysql"},
		ParameterGroupID: nil,
	}

	r.mapDatabaseToState(database, data)

	if data.ParameterGroupID.IsNull() {
		t.Fatal("ParameterGroupID = null, want preserved value pg-123")
	}
	if data.ParameterGroupID.ValueString() != "pg-123" {
		t.Errorf("ParameterGroupID = %v, want pg-123", data.ParameterGroupID.ValueString())
	}
}

func TestMapDatabaseToState_UpdatesParameterGroupIDWhenPresent(t *testing.T) {
	r := &DatabaseResource{}
	data := &DatabaseResourceModel{
		ParameterGroupID: types.StringValue("pg-old"),
	}
	newID := "pg-new"
	database := &client.DatabaseInstance{
		ID:               "db-1",
		Name:             "my-database",
		Status:           "running",
		Engine:           client.DatabaseEngine{ID: 1, Name: "mysql"},
		ParameterGroupID: &newID,
	}

	r.mapDatabaseToState(database, data)

	if data.ParameterGroupID.ValueString() != "pg-new" {
		t.Errorf("ParameterGroupID = %v, want pg-new", data.ParameterGroupID.ValueString())
	}
}

func TestMapDatabaseToState_EnginePrefersProviderType(t *testing.T) {
	r := &DatabaseResource{}
	data := &DatabaseResourceModel{}
	database := &client.DatabaseInstance{
		ID:       "db-1",
		Name:     "my-database",
		Status:   "running",
		Engine:   client.DatabaseEngine{ID: 1, Name: "mysql"},
		Provider: client.Provider{ID: 1, Name: "MySQL", Type: "mysql"},
	}

	r.mapDatabaseToState(database, data)

	if data.Engine.ValueString() != "mysql" {
		t.Errorf("Engine = %v, want mysql", data.Engine.ValueString())
	}
}

func TestMapDatabaseToState_EngineFallsBackToLowercasedName(t *testing.T) {
	r := &DatabaseResource{}
	data := &DatabaseResourceModel{}
	database := &client.DatabaseInstance{
		ID:     "db-1",
		Name:   "my-database",
		Status: "running",
		Engine: client.DatabaseEngine{ID: 1, Name: "mysql"},
	}

	r.mapDatabaseToState(database, data)

	if data.Engine.ValueString() != "mysql" {
		t.Errorf("Engine = %v, want mysql (lowercased fallback)", data.Engine.ValueString())
	}
}

func TestDatabaseNeedsStorageGrow(t *testing.T) {
	tests := []struct {
		name       string
		planned    types.Int64
		actualGB   int
		wantTarget int
		wantNeeded bool
	}{
		{name: "unconfigured", planned: types.Int64Null(), actualGB: 10, wantTarget: 0, wantNeeded: false},
		{name: "unknown", planned: types.Int64Unknown(), actualGB: 10, wantTarget: 0, wantNeeded: false},
		{name: "equal to provisioned", planned: types.Int64Value(10), actualGB: 10, wantTarget: 0, wantNeeded: false},
		{name: "below provisioned", planned: types.Int64Value(5), actualGB: 10, wantTarget: 0, wantNeeded: false},
		{name: "above provisioned", planned: types.Int64Value(50), actualGB: 10, wantTarget: 50, wantNeeded: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, needed := databaseNeedsStorageGrow(tt.planned, tt.actualGB)
			if target != tt.wantTarget || needed != tt.wantNeeded {
				t.Errorf("databaseNeedsStorageGrow(%v, %d) = (%d, %v), want (%d, %v)",
					tt.planned, tt.actualGB, target, needed, tt.wantTarget, tt.wantNeeded)
			}
		})
	}
}

// A DNS-only update must still refresh computed attributes from the API.
// Regression: mapDatabaseToState used to live inside `if hasChanges`, so a
// DNS-only change left status/username/updated_at unknown and Terraform
// rejected the apply with "invalid result object after apply".
func TestDatabaseUpdate_DnsOnlyChange_RefreshesComputedState(t *testing.T) {
	var getCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/dns"):
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet:
			getCalls++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"instance":{"id":"abc","name":"db1","status":"running","username":"root","updated_at":"2026-07-19T13:52:06+00:00"}}`))
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	c := client.New(client.Config{
		BaseURL:   srv.URL,
		APIToken:  "token",
		UserAgent: "test",
	})
	db, err := c.GetDatabase(context.Background(), "abc")
	if err != nil {
		t.Fatalf("GetDatabase: %v", err)
	}

	r := &DatabaseResource{client: c}
	data := &DatabaseResourceModel{}
	r.mapDatabaseToState(db, data)

	if data.Status.IsUnknown() || data.Status.ValueString() != "running" {
		t.Fatalf("status not refreshed: %#v", data.Status)
	}
	if data.Username.IsUnknown() || data.Username.ValueString() != "root" {
		t.Fatalf("username not refreshed: %#v", data.Username)
	}
	if getCalls != 1 {
		t.Fatalf("expected exactly 1 GET, got %d", getCalls)
	}
}
