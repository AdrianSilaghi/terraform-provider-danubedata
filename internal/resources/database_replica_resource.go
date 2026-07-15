package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &DatabaseReplicaResource{}
	_ resource.ResourceWithConfigure   = &DatabaseReplicaResource{}
	_ resource.ResourceWithImportState = &DatabaseReplicaResource{}
)

type DatabaseReplicaResource struct {
	client *client.Client
}

type DatabaseReplicaResourceModel struct {
	ID                   types.String   `tfsdk:"id"`
	DatabaseInstanceID   types.String   `tfsdk:"database_instance_id"`
	ReplicaIndex         types.Int64    `tfsdk:"replica_index"`
	Name                 types.String   `tfsdk:"name"`
	NodeID               types.String   `tfsdk:"node_id"`
	Endpoint             types.String   `tfsdk:"endpoint"`
	Status               types.String   `tfsdk:"status"`
	Ready                types.Bool     `tfsdk:"ready"`
	ReplicationStatus    types.String   `tfsdk:"replication_status"`
	SecondsBehindMaster  types.Int64    `tfsdk:"seconds_behind_master"`
	IsReplicationHealthy types.Bool     `tfsdk:"is_replication_healthy"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}

func NewDatabaseReplicaResource() resource.Resource {
	return &DatabaseReplicaResource{}
}

func (r *DatabaseReplicaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_replica"
}

func (r *DatabaseReplicaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a read replica for a DanubeData database instance. Use count or for_each to manage multiple replicas; add depends_on between replicas on the same instance to serialize creation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the form {database_instance_id}:{replica_index}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"database_instance_id": schema.StringAttribute{
				Description: "ID of the parent database instance.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"replica_index": schema.Int64Attribute{
				Description: "1-based index of this replica within the parent instance.",
				Computed:    true,
			},
			"name":                   schema.StringAttribute{Computed: true, Description: "Name of the replica node."},
			"node_id":                schema.StringAttribute{Computed: true, Description: "Internal node identifier."},
			"endpoint":               schema.StringAttribute{Computed: true, Description: "Connection endpoint for the replica."},
			"status":                 schema.StringAttribute{Computed: true, Description: "Current status of the replica."},
			"ready":                  schema.BoolAttribute{Computed: true, Description: "Whether the replica is ready to serve reads."},
			"replication_status":     schema.StringAttribute{Computed: true, Description: "Replication status (healthy, lagging, broken)."},
			"seconds_behind_master":  schema.Int64Attribute{Computed: true, Description: "Replication lag in seconds behind the master."},
			"is_replication_healthy": schema.BoolAttribute{Computed: true, Description: "Whether replication is healthy."},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *DatabaseReplicaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *DatabaseReplicaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseReplicaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	instanceID := data.DatabaseInstanceID.ValueString()

	tflog.Debug(ctx, "Adding database replica", map[string]interface{}{
		"database_instance_id": instanceID,
	})

	added, err := r.client.AddDatabaseReplicas(ctx, instanceID, client.AddDatabaseReplicasRequest{ReplicaCount: 1})
	if err != nil {
		resp.Diagnostics.AddError("Failed to add database replica", err.Error())
		return
	}
	if len(added) == 0 {
		resp.Diagnostics.AddError("Failed to add database replica", "API returned no replicas")
		return
	}

	newest := added[0]
	for _, rep := range added {
		if rep.ReplicaIndex > newest.ReplicaIndex {
			newest = rep
		}
	}

	if err := r.client.WaitForDatabaseReplicaReady(ctx, instanceID, newest.ReplicaIndex, createTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Database replica failed to become ready",
			fmt.Sprintf("Replica %s:%d did not become ready: %s", instanceID, newest.ReplicaIndex, err),
		)
		return
	}

	replica, err := r.client.FindDatabaseReplica(ctx, instanceID, newest.ReplicaIndex)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read replica after creation", err.Error())
		return
	}

	r.mapReplicaToState(instanceID, replica, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseReplicaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseReplicaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceID := data.DatabaseInstanceID.ValueString()
	idx := int(data.ReplicaIndex.ValueInt64())

	replica, err := r.client.FindDatabaseReplica(ctx, instanceID, idx)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read database replica", err.Error())
		return
	}

	r.mapReplicaToState(instanceID, replica, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseReplicaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable fields require replacement; Update is only invoked for computed-only
	// deltas (e.g., timeouts block). Read refreshes computed state on its own — so here we
	// preserve existing state rather than writing the plan back (which would contain
	// Unknown values for computed attributes and corrupt state).
	var data DatabaseReplicaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseReplicaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseReplicaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	instanceID := data.DatabaseInstanceID.ValueString()
	idx := int(data.ReplicaIndex.ValueInt64())

	if err := r.client.DeleteDatabaseReplica(ctx, instanceID, idx); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete database replica", err.Error())
		return
	}
}

func (r *DatabaseReplicaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected format: {database_instance_id}:{replica_index}
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: {database_instance_id}:{replica_index}, got: %s", req.ID),
		)
		return
	}
	idx, err := strconv.Atoi(parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Invalid replica_index in import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("database_instance_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("replica_index"), int64(idx))...)
}

func (r *DatabaseReplicaResource) mapReplicaToState(instanceID string, replica *client.DatabaseReplica, data *DatabaseReplicaResourceModel) {
	data.ID = types.StringValue(fmt.Sprintf("%s:%d", instanceID, replica.ReplicaIndex))
	data.DatabaseInstanceID = types.StringValue(instanceID)
	data.ReplicaIndex = types.Int64Value(int64(replica.ReplicaIndex))
	data.Name = types.StringValue(replica.Name)
	data.NodeID = types.StringValue(replica.NodeID)
	if replica.Endpoint != nil {
		data.Endpoint = types.StringValue(*replica.Endpoint)
	} else {
		data.Endpoint = types.StringNull()
	}
	data.Status = types.StringValue(replica.Status)
	data.Ready = types.BoolValue(replica.Ready)
	if replica.ReplicationStatus != nil {
		data.ReplicationStatus = types.StringValue(*replica.ReplicationStatus)
	} else {
		data.ReplicationStatus = types.StringNull()
	}
	if replica.SecondsBehindMaster != nil {
		data.SecondsBehindMaster = types.Int64Value(int64(*replica.SecondsBehindMaster))
	} else {
		data.SecondsBehindMaster = types.Int64Null()
	}
	data.IsReplicationHealthy = types.BoolValue(replica.IsReplicationHealthy)
}
