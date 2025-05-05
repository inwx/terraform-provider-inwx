package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/inwx/terraform-provider-inwx/internal/api"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var validRenewalModes = []string{"AUTORENEW", "AUTODELETE", "AUTOEXPIRE"}

type domainResource struct {
	client *api.Client
}

// Ensure the implementation satisfies the Terraform resource interface.
var _ resource.Resource = &domainResource{}
var _ resource.ResourceWithConfigure = &domainResource{}

// NewDomainResource returns a new instance of the domain resource.
func NewDomainResource() resource.Resource {
	return &domainResource{}
}

// Metadata defines the resource type name.
func (r *domainResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "inwx_domain"
}

// Configure sets the client for the resource.
func (r *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema defines the attributes of the resource.
func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage a domain using the INWX API.

## Caveats

### Extra Data

When extra data is set, e.g. ` + "`" + `WHOIS-PROTECTION` + "`" + `, our system sometimes adds other readonly extra data to the domain.
In this example ` + "`" + `WHOIS-CURRENCY` + "`" + ` is added to the domain. Terraform cannot manage this extra data, so it is recommended
to ignore these side effects explicitly as they occur:

` + "```" + `terraform
resource "inwx_domain" "example_com" {
  // ...
  extra_data = {
    "WHOIS-PROTECTION": "1"
  }

  lifecycle {
    ignore_changes = [
      extra_data["WHOIS-CURRENCY"], // ignore WHOIS-CURRENCY
    ]
  }
}
` + "```" + `

`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the domain.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nameservers": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of nameservers for the domain.",
			},
			"period": schema.StringAttribute{
				Description: "Registration period of the domain.",
				Required:    true,
			},
			"renewal_mode": schema.StringAttribute{
				Description: fmt.Sprintf("Renewal mode of the domain. One of: %s", strings.Join(validRenewalModes, ", ")),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(validRenewalModes...),
				},
				Default:  stringdefault.StaticString("AUTORENEW"),
				Computed: true,
			},
			"transfer_lock": schema.BoolAttribute{
				Description: "Whether the domain transfer lock should be enabled.",
				Optional:    true,
				Default:     booldefault.StaticBool(true),
				Computed:    true,
			},
			"contacts": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Contacts of the domain, depending on tld there might be different contact types that are required.",
				Attributes: map[string]schema.Attribute{
					"registrant": schema.Int64Attribute{
						Description: "ID of the registrant contact is always required.",
						Required:    true,
					},
					"admin": schema.Int64Attribute{
						Description: "ID of the admin contact might be optional depending on the tld. If optional will not be visible in API after creation.",
						Required:    true,
					},
					"tech": schema.Int64Attribute{
						Description: "ID of the tech contact might be optional depending on the tld. If optional will not be visible in API after creation.",
						Required:    true,
					},
					"billing": schema.Int64Attribute{
						Description: "ID of the billing contact might be optional depending on the tld. If optional will not be visible in API after creation.",
						Required:    true,
					},
				},
			},
			"extra_data": schema.MapAttribute{
				Description: "Extra data needed for certain jurisdictions.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Create handles the creation of the resource.
func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data domainResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"domain":       data.Name.ValueString(),
		"ns":           data.Nameservers.ElementsAs(ctx, &[]string{}, false),
		"period":       data.Period.ValueString(),
		"registrant":   data.Contacts.Registrant.ValueInt64(),
		"admin":        data.Contacts.Admin.ValueInt64(),
		"tech":         data.Contacts.Tech.ValueInt64(),
		"billing":      data.Contacts.Billing.ValueInt64(),
		"transferLock": data.TransferLock.ValueBool(),
		"renewalMode":  data.RenewalMode.ValueString(),
		"extData":      data.ExtraData.ElementsAs(ctx, &map[string]string{}, false),
	}

	call, err := r.client.Call(ctx, "domain.create", parameters)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create domain", err.Error())
		return
	}

	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()))
		return
	}

	data.ID = types.StringValue(data.Name.ValueString())
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (c domainContacts) Equal(other domainContacts) bool {
	return c.Registrant.Equal(other.Registrant) &&
		c.Admin.Equal(other.Admin) &&
		c.Tech.Equal(other.Tech) &&
		c.Billing.Equal(other.Billing)
}

// Read handles reading the resource data.
func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainResourceModel

	// Retrieve the current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the INWX API to fetch the current domain info
	parameters := map[string]interface{}{
		"domain": state.Name.ValueString(),
		"wide":   2,
	}

	call, err := r.client.Call(ctx, "domain.info", parameters)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch domain info", err.Error())
		return
	}

	if call.Code() != api.COMMAND_SUCCESSFUL {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Unexpected response code: %s", call.ApiError()),
		)
		return
	}

	// Process API response
	resData := call["resData"].(map[string]interface{})

	// Set state values from the response
	state.Name = types.StringValue(resData["domain"].(string))
	state.Period = types.StringValue(resData["period"].(string))
	state.RenewalMode = types.StringValue(resData["renewalMode"].(string))
	state.TransferLock = types.BoolValue(resData["transferLock"].(bool))

	// Convert nameservers to []attr.Value
	nameservers := resData["ns"].([]interface{})
	nsValues := make([]attr.Value, len(nameservers))
	for i, ns := range nameservers {
		nsValues[i] = types.StringValue(ns.(string))
	}
	state.Nameservers = types.ListValueMust(types.StringType, nsValues)

	// Convert contacts
	// Extract contacts from API response
	contacts := resData["contacts"].(map[string]interface{})

	// Map API response to domainContacts
	contactValues := domainContacts{
		Registrant: types.Int64Value(int64(contacts["registrant"].(float64))),
		Admin:      types.Int64Value(int64(contacts["admin"].(float64))),
		Tech:       types.Int64Value(int64(contacts["tech"].(float64))),
		Billing:    types.Int64Value(int64(contacts["billing"].(float64))),
	}

	// Assign contacts to state
	state.Contacts = contactValues

	// Convert extra_data to map[string]string
	// Convert extra_data to map[string]attr.Value
	extraData := resData["extData"].(map[string]interface{})
	extraDataMap := make(map[string]attr.Value, len(extraData))
	for k, v := range extraData {
		extraDataMap[k] = types.StringValue(v.(string))
	}
	state.ExtraData = types.MapValueMust(types.StringType, extraDataMap)

	// Save the updated state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update handles updating the resource data.
func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainResourceModel
	var state domainResourceModel

	// Retrieve the planned state and current state
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize parameters for the API call
	parameters := map[string]interface{}{
		"domain": state.Name.ValueString(),
	}

	// Detect changes and populate the parameters accordingly
	if !plan.Nameservers.Equal(state.Nameservers) {
		parameters["ns"] = plan.Nameservers.ElementsAs(ctx, &[]string{}, false)
	}

	if !plan.Period.Equal(state.Period) {
		parameters["period"] = plan.Period.ValueString()
	}

	if !plan.RenewalMode.Equal(state.RenewalMode) {
		parameters["renewalMode"] = plan.RenewalMode.ValueString()
	}

	if !plan.TransferLock.Equal(state.TransferLock) {
		parameters["transferLock"] = plan.TransferLock.ValueBool()
	}

	if !plan.Contacts.Equal(state.Contacts) {
		contacts := plan.Contacts
		parameters["registrant"] = contacts.Registrant.ValueInt64()
		parameters["admin"] = contacts.Admin.ValueInt64()
		parameters["tech"] = contacts.Tech.ValueInt64()
		parameters["billing"] = contacts.Billing.ValueInt64()
	}

	if !plan.ExtraData.Equal(state.ExtraData) {
		parameters["extData"] = plan.ExtraData.ElementsAs(ctx, &map[string]string{}, false)
	}

	// Call the INWX API to update the domain
	call, err := r.client.Call(ctx, "domain.update", parameters)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update domain", err.Error())
		return
	}

	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError(
			"API error",
			fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		)
		return
	}

	// Save the updated state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete handles resource deletion.
func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data domainResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"domain": data.Name.ValueString(),
	}

	call, err := r.client.Call(ctx, "domain.delete", parameters)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete domain", err.Error())
		return
	}

	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()))
	}
}

type domainResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	Nameservers  types.List     `tfsdk:"nameservers"`
	Period       types.String   `tfsdk:"period"`
	RenewalMode  types.String   `tfsdk:"renewal_mode"`
	TransferLock types.Bool     `tfsdk:"transfer_lock"`
	Contacts     domainContacts `tfsdk:"contacts"`
	ExtraData    types.Map      `tfsdk:"extra_data"`
}

type domainContacts struct {
	Registrant types.Int64 `tfsdk:"registrant"`
	Admin      types.Int64 `tfsdk:"admin"`
	Tech       types.Int64 `tfsdk:"tech"`
	Billing    types.Int64 `tfsdk:"billing"`
}
