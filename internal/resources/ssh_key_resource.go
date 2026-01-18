package resources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &SshKeyResource{}
	_ resource.ResourceWithConfigure   = &SshKeyResource{}
	_ resource.ResourceWithImportState = &SshKeyResource{}
)

type SshKeyResource struct {
	client *client.Client
}

type SshKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PublicKey   types.String `tfsdk:"public_key"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewSshKeyResource() resource.Resource {
	return &SshKeyResource{}
}

func (r *SshKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *SshKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData SSH key for VPS authentication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the SSH key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "A descriptive name for the SSH key.",
				Required:    true,
			},
			"public_key": schema.StringAttribute{
				Description: "The SSH public key in OpenSSH format (e.g., 'ssh-rsa AAAA...' or 'ssh-ed25519 AAAA...').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fingerprint": schema.StringAttribute{
				Description: "The SHA256 fingerprint of the SSH key.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the SSH key was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the SSH key was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *SshKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SshKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating SSH key", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	createReq := client.CreateSshKeyRequest{
		Name:      data.Name.ValueString(),
		PublicKey: data.PublicKey.ValueString(),
	}

	sshKey, err := r.client.CreateSshKey(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SSH key", err.Error())
		return
	}

	r.mapSshKeyToState(sshKey, &data)

	tflog.Info(ctx, "SSH key created", map[string]interface{}{
		"id":   sshKey.ID,
		"name": sshKey.Name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SshKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sshKey, err := r.client.GetSshKey(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SSH key", err.Error())
		return
	}

	r.mapSshKeyToState(sshKey, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SshKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// SSH keys can only be replaced, not updated in place
	// The public_key has RequiresReplace, so only name changes come here
	// The API doesn't support updating SSH keys, so we just refresh state
	sshKey, err := r.client.GetSshKey(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SSH key after update", err.Error())
		return
	}

	r.mapSshKeyToState(sshKey, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SshKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting SSH key", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	err := r.client.DeleteSshKey(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete SSH key", err.Error())
		return
	}
}

func (r *SshKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SshKeyResource) mapSshKeyToState(sshKey *client.SshKey, data *SshKeyResourceModel) {
	data.ID = types.StringValue(sshKey.ID)
	data.Name = types.StringValue(sshKey.Name)
	data.PublicKey = types.StringValue(sshKey.PublicKey)
	data.Fingerprint = types.StringValue(sshKey.Fingerprint)
	data.CreatedAt = types.StringValue(sshKey.CreatedAt)
	data.UpdatedAt = types.StringValue(sshKey.UpdatedAt)
}
