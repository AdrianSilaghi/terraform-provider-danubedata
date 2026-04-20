package resources

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &StaticSiteDomainResource{}
	_ resource.ResourceWithConfigure   = &StaticSiteDomainResource{}
	_ resource.ResourceWithImportState = &StaticSiteDomainResource{}
)

type StaticSiteDomainResource struct {
	client *client.Client
}

type StaticSiteDomainResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	StaticSiteID       types.String `tfsdk:"static_site_id"`
	DomainID           types.Int64  `tfsdk:"domain_id"`
	Domain             types.String `tfsdk:"domain"`
	Type               types.String `tfsdk:"type"`
	Status             types.String `tfsdk:"status"`
	VerificationRecord types.String `tfsdk:"verification_record"`
	VerifiedAt         types.String `tfsdk:"verified_at"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func NewStaticSiteDomainResource() resource.Resource {
	return &StaticSiteDomainResource{}
}

func (r *StaticSiteDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_site_domain"
}

func (r *StaticSiteDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom domain attached to a DanubeData static site. After the resource is created the domain is in `pending` status with a `verification_record` to add to DNS; verification is triggered out-of-band via `danube pages domains verify` once the DNS record is in place.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the form {static_site_id}:{domain_id}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"static_site_id": schema.StringAttribute{
				Description: "ID of the parent static site.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_id": schema.Int64Attribute{
				Description: "Numeric ID of the domain attachment.",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "The custom domain (e.g., www.example.com).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Domain type: default or custom.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the domain (pending, active, failed).",
				Computed:    true,
			},
			"verification_record": schema.StringAttribute{
				Description: "DNS record to configure for verification (CNAME). Empty after verification succeeds.",
				Computed:    true,
			},
			"verified_at": schema.StringAttribute{
				Description: "Timestamp when the domain was verified, if any.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the domain attachment was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the domain attachment was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *StaticSiteDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StaticSiteDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StaticSiteDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := data.StaticSiteID.ValueString()
	tflog.Debug(ctx, "Adding static site domain", map[string]interface{}{
		"static_site_id": siteID,
		"domain":         data.Domain.ValueString(),
	})

	domain, err := r.client.AddStaticSiteDomain(ctx, siteID, client.AddStaticSiteDomainRequest{
		Domain: data.Domain.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to add static site domain", err.Error())
		return
	}

	r.mapDomainToState(siteID, domain, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticSiteDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StaticSiteDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := data.StaticSiteID.ValueString()
	domain, err := r.client.FindStaticSiteDomain(ctx, siteID, data.Domain.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read static site domain", err.Error())
		return
	}

	r.mapDomainToState(siteID, domain, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticSiteDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Domain attachments are immutable; all user-visible fields require replacement.
	// Preserve existing state rather than writing the plan.
	var data StaticSiteDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticSiteDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StaticSiteDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := data.StaticSiteID.ValueString()
	domainID := fmt.Sprintf("%d", data.DomainID.ValueInt64())

	if err := r.client.DeleteStaticSiteDomain(ctx, siteID, domainID); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete static site domain", err.Error())
		return
	}
}

func (r *StaticSiteDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected format: {static_site_id}:{domain}
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: {static_site_id}:{domain}, got: %s", req.ID),
		)
		return
	}
	// Set id to the composite import string as a placeholder so framework doesn't reject
	// the state before Read runs; mapDomainToState rewrites it with the numeric form.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("static_site_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[1])...)
}

func (r *StaticSiteDomainResource) mapDomainToState(siteID string, domain *client.StaticSiteDomain, data *StaticSiteDomainResourceModel) {
	data.ID = types.StringValue(fmt.Sprintf("%s:%d", siteID, domain.ID))
	data.StaticSiteID = types.StringValue(siteID)
	data.DomainID = types.Int64Value(int64(domain.ID))
	data.Domain = types.StringValue(domain.Domain)
	data.Type = types.StringValue(domain.Type)
	data.Status = types.StringValue(domain.Status)
	if domain.VerificationRecord != nil {
		data.VerificationRecord = types.StringValue(*domain.VerificationRecord)
	} else {
		data.VerificationRecord = types.StringNull()
	}
	if domain.VerifiedAt != nil {
		data.VerifiedAt = types.StringValue(*domain.VerifiedAt)
	} else {
		data.VerifiedAt = types.StringNull()
	}
	data.CreatedAt = types.StringValue(domain.CreatedAt)
	data.UpdatedAt = types.StringValue(domain.UpdatedAt)
}
