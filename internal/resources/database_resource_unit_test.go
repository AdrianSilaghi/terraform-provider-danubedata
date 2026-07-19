package resources

import (
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
