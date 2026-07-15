package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	DomainID           types.String `tfsdk:"domain_id"`
	Domain             types.String `tfsdk:"domain"`
	VerificationStatus types.String `tfsdk:"verification_status"`
	TLSStatus          types.String `tfsdk:"tls_status"`
	DeploymentStatus   types.String `tfsdk:"deployment_status"`
	IsPrimary          types.Bool   `tfsdk:"is_primary"`
	DNSInstructions    types.Object `tfsdk:"dns_instructions"`
	CreatedAt          types.String `tfsdk:"created_at"`
}

// staticSiteDomainDNSInstructionsAttrTypes describes the object type of the dns_instructions attribute.
var staticSiteDomainDNSInstructionsAttrTypes = map[string]attr.Type{
	"record_type":  types.StringType,
	"record_name":  types.StringType,
	"record_value": types.StringType,
	"instructions": types.StringType,
}

func NewStaticSiteDomainResource() resource.Resource {
	return &StaticSiteDomainResource{}
}

func (r *StaticSiteDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_site_domain"
}

func (r *StaticSiteDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom domain attached to a DanubeData static site. After the resource is created the domain is in `pending` verification status; add the DNS record described in `dns_instructions` to prove ownership, then trigger verification out-of-band via `danube pages domains verify` once the record is in place.",
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
			"domain_id": schema.StringAttribute{
				Description: "ID of the domain attachment.",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "The custom domain (e.g., www.example.com).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"verification_status": schema.StringAttribute{
				Description: "DNS ownership verification status (pending, verifying, verified, failed).",
				Computed:    true,
			},
			"tls_status": schema.StringAttribute{
				Description: "TLS certificate provisioning status for the domain (pending, provisioning, active, failed).",
				Computed:    true,
			},
			"deployment_status": schema.StringAttribute{
				Description: "Status of routing the domain to the site's active deployment (pending, deploying, active, failed).",
				Computed:    true,
			},
			"is_primary": schema.BoolAttribute{
				Description: "Whether this is the primary domain for the site.",
				Computed:    true,
			},
			"dns_instructions": schema.SingleNestedAttribute{
				Description: "DNS record to add for ownership verification.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"record_type": schema.StringAttribute{
						Description: "DNS record type (e.g., TXT).",
						Computed:    true,
					},
					"record_name": schema.StringAttribute{
						Description: "DNS record name.",
						Computed:    true,
					},
					"record_value": schema.StringAttribute{
						Description: "DNS record value.",
						Computed:    true,
					},
					"instructions": schema.StringAttribute{
						Description: "Human-readable instructions for configuring the record.",
						Computed:    true,
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the domain attachment was created.",
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

	r.mapDomainToState(siteID, domain, &data, &resp.Diagnostics)
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

	r.mapDomainToState(siteID, domain, &data, &resp.Diagnostics)
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
	domainID := data.DomainID.ValueString()

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
	// the state before Read runs; mapDomainToState rewrites it once the domain UUID is known.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("static_site_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[1])...)
}

func (r *StaticSiteDomainResource) mapDomainToState(siteID string, domain *client.StaticSiteDomain, data *StaticSiteDomainResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(fmt.Sprintf("%s:%s", siteID, domain.ID))
	data.StaticSiteID = types.StringValue(siteID)
	data.DomainID = types.StringValue(domain.ID)
	data.Domain = types.StringValue(domain.Domain)
	data.VerificationStatus = types.StringValue(domain.VerificationStatus)
	data.TLSStatus = types.StringValue(domain.TLSStatus)
	data.DeploymentStatus = types.StringValue(domain.DeploymentStatus)
	data.IsPrimary = types.BoolValue(domain.IsPrimary)

	dnsInstructions, dnsDiags := types.ObjectValue(
		staticSiteDomainDNSInstructionsAttrTypes,
		map[string]attr.Value{
			"record_type":  types.StringValue(domain.DNSInstructions.RecordType),
			"record_name":  types.StringValue(domain.DNSInstructions.RecordName),
			"record_value": types.StringValue(domain.DNSInstructions.RecordValue),
			"instructions": types.StringValue(domain.DNSInstructions.Instructions),
		},
	)
	diags.Append(dnsDiags...)
	data.DNSInstructions = dnsInstructions

	data.CreatedAt = types.StringValue(domain.CreatedAt)
}
