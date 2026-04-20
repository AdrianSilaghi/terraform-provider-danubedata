package resources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &StaticSiteResource{}
	_ resource.ResourceWithConfigure   = &StaticSiteResource{}
	_ resource.ResourceWithImportState = &StaticSiteResource{}
)

type StaticSiteResource struct {
	client *client.Client
}

type StaticSiteResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	TeamID              types.Int64  `tfsdk:"team_id"`
	Name                types.String `tfsdk:"name"`
	Slug                types.String `tfsdk:"slug"`
	URL                 types.String `tfsdk:"url"`
	OutputDirectory     types.String `tfsdk:"output_directory"`
	Status              types.String `tfsdk:"status"`
	CurrentDeploymentID types.Int64  `tfsdk:"current_deployment_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func NewStaticSiteResource() resource.Resource {
	return &StaticSiteResource{}
}

func (r *StaticSiteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_site"
}

func (r *StaticSiteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData static site (pages). Deployments are triggered out-of-band via the CLI or CI/CD; this resource manages only the site container, not its content.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the static site.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.Int64Attribute{
				Description: "ID of the team that owns this site.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the static site.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "URL slug for the site.",
				Computed:    true,
			},
			"url": schema.StringAttribute{
				Description: "Default URL of the deployed site.",
				Computed:    true,
			},
			"output_directory": schema.StringAttribute{
				Description: "Build output directory served as the site root.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the site.",
				Computed:    true,
			},
			"current_deployment_id": schema.Int64Attribute{
				Description: "ID of the currently-active deployment, if any.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the site was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the site was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *StaticSiteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StaticSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StaticSiteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating static site", map[string]interface{}{
		"name":    data.Name.ValueString(),
		"team_id": data.TeamID.ValueInt64(),
	})

	site, err := r.client.CreateStaticSite(ctx, int(data.TeamID.ValueInt64()), client.CreateStaticSiteRequest{
		Name: data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create static site", err.Error())
		return
	}

	r.mapSiteToState(site, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StaticSiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	site, err := r.client.GetStaticSite(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read static site", err.Error())
		return
	}

	r.mapSiteToState(site, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Static sites are immutable through this API; all user-visible fields require
	// replacement. Preserve existing state rather than writing the plan (which may
	// contain Unknown computed fields).
	var data StaticSiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StaticSiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteStaticSite(ctx, data.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete static site", err.Error())
		return
	}
}

func (r *StaticSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *StaticSiteResource) mapSiteToState(site *client.StaticSite, data *StaticSiteResourceModel) {
	data.ID = types.StringValue(fmt.Sprintf("%d", site.ID))
	data.TeamID = types.Int64Value(int64(site.TeamID))
	data.Name = types.StringValue(site.Name)
	data.Slug = types.StringValue(site.Slug)
	data.URL = types.StringValue(site.URL)
	if site.OutputDirectory != nil {
		data.OutputDirectory = types.StringValue(*site.OutputDirectory)
	} else {
		data.OutputDirectory = types.StringNull()
	}
	data.Status = types.StringValue(site.Status)
	if site.CurrentDeploymentID != nil {
		data.CurrentDeploymentID = types.Int64Value(int64(*site.CurrentDeploymentID))
	} else {
		data.CurrentDeploymentID = types.Int64Null()
	}
	data.CreatedAt = types.StringValue(site.CreatedAt)
	data.UpdatedAt = types.StringValue(site.UpdatedAt)
}
