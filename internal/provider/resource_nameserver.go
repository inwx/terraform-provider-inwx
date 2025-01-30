package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/inwx/terraform-provider-inwx/internal/api"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NameserverResource{}
var _ resource.ResourceWithImportState = &NameserverResource{}

func NewNameserverResource() resource.Resource {
	return &NameserverResource{}
}

type NameserverResource struct {
	client *api.Client
}

func (r *NameserverResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "inwx_nameserver"
}

func (r *NameserverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	validTypes := []string{"MASTER", "SLAVE"}
	validUrlRedirectTypes := []string{"HEADER301", "HEADER302", "FRAME"}

	resp.Schema = schema.Schema{
		Description: "Provides a INWX nameserver zone resource on the anycast nameserver network (50+ locations worldwide). Needed if you use INWX nameservers for inwx_domain. Use inwx_nameserver_record to create records in the zone.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "Domain name",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: fmt.Sprintf("Type of the nameserver. One of: %s", strings.Join(validTypes, ", ")),
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(validTypes...),
				},
			},
			"nameservers": schema.ListAttribute{
				Description: "List of nameservers",
				ElementType: types.StringType,
				Required:    true,
			},
			"master_ip": schema.StringAttribute{
				Description: "Master IP address",
				Optional:    true,
			},
			"web": schema.StringAttribute{
				Description: "Web nameserver entry",
				Optional:    true,
			},
			"mail": schema.StringAttribute{
				Description: "Mail nameserver entry",
				Optional:    true,
			},
			"soa_mail": schema.StringAttribute{
				Description: "Email address for SOA record",
				Optional:    true,
			},
			"url_redirect_type": schema.StringAttribute{
				Description: fmt.Sprintf("Type of the URL redirection. One of: %s", strings.Join(validUrlRedirectTypes, ", ")),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(validUrlRedirectTypes...),
				},
			},
			"url_redirect_title": schema.StringAttribute{
				Description: "Title of the frame redirection",
				Optional:    true,
			},
			"url_redirect_description": schema.StringAttribute{
				Description: "Description of the frame redirection",
				Optional:    true,
			},
			"url_redirect_fav_icon": schema.StringAttribute{
				Description: "FavIcon of the frame redirection",
				Optional:    true,
			},
			"url_redirect_keywords": schema.StringAttribute{
				Description: "Keywords of the frame redirection",
				Optional:    true,
			},
			"testing": schema.BoolAttribute{
				Description: "Execute command in testing mode",
				Optional:    true,
			},
			"ignore_existing": schema.BoolAttribute{
				Description: "Ignore existing",
				Optional:    true,
			},
		},
	}
}

func (r *NameserverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *NameserverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NameserverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"domain": plan.Domain.ValueString(),
		"type":   plan.Type.ValueString(),
	}

	if !plan.Nameservers.IsNull() {
		parameters["ns"] = plan.Nameservers.Elements()
	}
	if !plan.MasterIp.IsNull() {
		parameters["masterIp"] = plan.MasterIp.ValueString()
	}
	if !plan.Web.IsNull() {
		parameters["web"] = plan.Web.ValueString()
	}
	if !plan.Mail.IsNull() {
		parameters["mail"] = plan.Mail.ValueString()
	}
	if !plan.UrlRedirectType.IsNull() {
		parameters["urlRedirectType"] = plan.UrlRedirectType.ValueString()
	}
	if !plan.UrlRedirectTitle.IsNull() {
		parameters["urlRedirectTitle"] = plan.UrlRedirectTitle.ValueString()
	}
	if !plan.UrlRedirectDescription.IsNull() {
		parameters["urlRedirectDescription"] = plan.UrlRedirectDescription.ValueString()
	}
	if !plan.UrlRedirectFavIcon.IsNull() {
		parameters["urlRedirectFavIcon"] = plan.UrlRedirectFavIcon.ValueString()
	}
	if !plan.UrlRedirectKeywords.IsNull() {
		parameters["urlRedirectKeywords"] = plan.UrlRedirectKeywords.ValueString()
	}
	if !plan.Testing.IsNull() {
		parameters["testing"] = plan.Testing.ValueBool()
	}
	if !plan.IgnoreExisting.IsNull() {
		parameters["ignoreExisting"] = plan.IgnoreExisting.ValueBool()
	}

	call, err := r.client.Call(ctx, "nameserver.create", parameters)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not add nameserver record: %s", err))
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()))
		return
	}

	resData := call["resData"].(map[string]any)
	roID := strconv.Itoa(int(resData["roId"].(float64)))
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.Domain.ValueString(), roID))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NameserverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NameserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"domain": state.Domain.ValueString(),
	}

	call, err := r.client.Call(ctx, "nameserver.info", parameters)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not read nameserver record: %s", err))
		return
	}

	if resData, ok := call["resData"].(map[string]any); ok {
		domain := resData["domain"].(string)
		state.Domain = types.StringValue(domain)

		if t, ok := resData["type"]; ok {
			state.Type = types.StringValue(t.(string))
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *NameserverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No-op Update: Since this resource doesn't support updates, we just read the existing state.
	resp.Diagnostics.AddWarning(
		"No Update Support",
		"This resource does not support updates. To make changes, please delete and recreate the resource.",
	)

	// Read the current state into the response's State.
	var state NameserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the state back, ensuring nothing has changed.
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *NameserverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NameserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"domain": state.Domain.ValueString(),
	}

	if !state.Testing.IsNull() {
		parameters["testing"] = state.Testing.ValueBool()
	}

	err := r.client.CallNoResponseBody(ctx, "nameserver.delete", parameters)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not delete nameserver record: %s", err))
		return
	}
}

func (r *NameserverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	domain, id, err := resourceNameserverParseId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing ID", fmt.Sprintf("Could not parse ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), domain)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ro_id"), id)...)
}

type NameserverResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Domain                 types.String `tfsdk:"domain"`
	Type                   types.String `tfsdk:"type"`
	Nameservers            types.List   `tfsdk:"nameservers"`
	MasterIp               types.String `tfsdk:"master_ip"`
	Web                    types.String `tfsdk:"web"`
	Mail                   types.String `tfsdk:"mail"`
	SoaMail                types.String `tfsdk:"soa_mail"`
	UrlRedirectType        types.String `tfsdk:"url_redirect_type"`
	UrlRedirectTitle       types.String `tfsdk:"url_redirect_title"`
	UrlRedirectDescription types.String `tfsdk:"url_redirect_description"`
	UrlRedirectFavIcon     types.String `tfsdk:"url_redirect_fav_icon"`
	UrlRedirectKeywords    types.String `tfsdk:"url_redirect_keywords"`
	Testing                types.Bool   `tfsdk:"testing"`
	IgnoreExisting         types.Bool   `tfsdk:"ignore_existing"`
}

func resourceNameserverParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected attribute1:attribute2", id)
	}

	return parts[0], parts[1], nil
}
