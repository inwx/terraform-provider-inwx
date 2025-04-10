package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/inwx/terraform-provider-inwx/internal/api"
	"reflect"
	"strconv"
)

type domainContactResource struct {
	client *api.Client
}

// Ensure the implementation satisfies the Terraform resource interface.
var _ resource.Resource = &domainResource{}
var _ resource.ResourceWithConfigure = &domainResource{}

func NewDomainContactResource() resource.Resource {
	return &domainContactResource{}
}

func (r *domainContactResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "inwx_domain_contact"
}

// Configure sets the client for the resource.
func (r *domainContactResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *domainContactResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a INWX domain contact resource. Needed for inwx_domain.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of contact. One of: ORG, PERSON, ROLE.",
				Validators: []validator.String{
					stringvalidator.OneOf("ORG", "PERSON", "ROLE"),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "First and last name of the contact.",
			},
			"organization": schema.StringAttribute{
				Optional:    true,
				Description: "The legal name of the organization. Required for types other than PERSON.",
			},
			"street_address": schema.StringAttribute{
				Required:    true,
				Description: "Street address of the contact.",
			},
			"city": schema.StringAttribute{
				Required:    true,
				Description: "City of the contact.",
			},
			"postal_code": schema.StringAttribute{
				Required:    true,
				Description: "Postal code/ZIP code of the contact.",
			},
			"state_province": schema.StringAttribute{
				Optional:    true,
				Description: "State or province name of the contact.",
			},
			"country_code": schema.StringAttribute{
				Required:    true,
				Description: "Country code of the contact. Must be two characters.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 2),
				},
			},
			"phone_number": schema.StringAttribute{
				Required:    true,
				Description: "Phone number of the contact.",
			},
			"fax": schema.StringAttribute{
				Optional:    true,
				Description: "Fax number of the contact.",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "Contact email address.",
			},
			"remarks": schema.StringAttribute{
				Optional:    true,
				Description: "Custom description of the contact. Max length is 255 characters.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"whois_protection": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether whois protection for the contact should be enabled.",
				Default:     booldefault.StaticBool(true),
				Computed:    true,
			},
		},
	}
}

// DomainContactModel represents the data structure for a domain contact resource
type domainContactModel struct {
	ID              types.String `json:"id"`               // Unique identifier for the contact
	Type            types.String `json:"type"`             // Type of the contact (e.g., ORG, PERSON, ROLE)
	Name            types.String `json:"name"`             // First and last name of the contact
	Organization    types.String `json:"organization"`     // Organization name (optional)
	StreetAddress   types.String `json:"street_address"`   // Street address of the contact
	City            types.String `json:"city"`             // City of the contact
	PostalCode      types.String `json:"postal_code"`      // Postal/ZIP code of the contact
	StateProvince   types.String `json:"state_province"`   // State/Province (optional)
	CountryCode     types.String `json:"country_code"`     // Country code (ISO 3166-1 alpha-2)
	PhoneNumber     types.String `json:"phone_number"`     // Contact's phone number
	Fax             types.String `json:"fax"`              // Fax number (optional)
	Email           types.String `json:"email"`            // Contact's email address
	Remarks         types.String `json:"remarks"`          // Remarks or custom description (optional)
	WhoisProtection types.Bool   `json:"whois_protection"` // Indicates if WHOIS protection is enabled
}

func (m domainContactModel) ToModel(data map[string]interface{}) domainContactModel {
	return domainContactModel{
		ID:              types.StringValue(data["id"].(string)),
		Type:            types.StringValue(data["type"].(string)),
		Name:            types.StringValue(data["name"].(string)),
		Organization:    optionalString(data, "org"),
		StreetAddress:   optionalString(data, "street"),
		City:            types.StringValue(data["city"].(string)),
		PostalCode:      types.StringValue(data["pc"].(string)),
		StateProvince:   optionalString(data, "sp"),
		CountryCode:     types.StringValue(data["cc"].(string)),
		PhoneNumber:     types.StringValue(data["voice"].(string)),
		Fax:             optionalString(data, "fax"),
		Email:           types.StringValue(data["email"].(string)),
		Remarks:         optionalString(data, "remarks"),
		WhoisProtection: types.BoolValue(data["protection"].(bool)),
	}
}

func optionalString(data map[string]interface{}, key string) types.String {
	if value, exists := data[key]; exists && value != nil {
		return types.StringValue(value.(string))
	}
	return types.StringNull()
}

func (r *domainContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data domainContactModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"type":       data.Type.ValueString(),
		"name":       data.Name.ValueString(),
		"street":     data.StreetAddress.ValueString(),
		"city":       data.City.ValueString(),
		"pc":         data.PostalCode.ValueString(),
		"cc":         data.CountryCode.ValueString(),
		"voice":      data.PhoneNumber.ValueString(),
		"email":      data.Email.ValueString(),
		"protection": data.WhoisProtection.ValueBool(),
	}
	if !data.Organization.IsNull() {
		parameters["org"] = data.Organization.ValueString()
	}
	if !data.StateProvince.IsNull() {
		parameters["sp"] = data.StateProvince.ValueString()
	}
	if !data.Fax.IsNull() {
		parameters["fax"] = data.Fax.ValueString()
	}
	if !data.Remarks.IsNull() {
		parameters["remarks"] = data.Remarks.ValueString()
	}

	call, err := r.client.Call(ctx, "contact.create", parameters)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create contact: %s", err.Error()))
		return
	}

	rawID := call["resData"].(map[string]interface{})["id"]
	id, diags := resolveID(rawID)
	if diags != nil {
		resp.Diagnostics.AddError("ID Error", fmt.Sprintf("Failed to resolve contact ID: %s", diags.Errors()))
		return
	}
	data.ID = types.StringValue(id)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *domainContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data domainContactModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	contactID, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("ID Conversion Error", fmt.Sprintf("Failed to convert contact ID to int: %s", err.Error()))
		return
	}

	parameters := map[string]interface{}{
		"id":   contactID,
		"wide": 2,
	}
	call, err := r.client.Call(ctx, "contact.info", parameters)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read contact: %s", err.Error()))
		return
	}

	contact, diags := expandContactFromInfoResponse(call["resData"].(map[string]interface{})["contact"].(map[string]interface{}))
	if diags != nil {
		resp.Diagnostics.AddError("Contact Exapnsion Error", fmt.Sprintf("Failed to expand contact info from response: %s", diags.Errors()))
		return
	}
	data = contact.ToModel(call["resData"].(map[string]interface{})["contact"].(map[string]interface{}))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *domainContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainContactModel
	var state domainContactModel

	// Retrieve plan and current state
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

	// Compare changes and prepare update parameters
	updateParameters := map[string]interface{}{
		"id": state.ID.ValueString(),
	}

	if !plan.Type.Equal(state.Type) {
		resp.Diagnostics.AddError(
			"Unsupported Update",
			"The `type` field cannot be updated after resource creation.",
		)
		return
	}
	if !plan.Name.Equal(state.Name) {
		updateParameters["name"] = plan.Name.ValueString()
	}
	if !plan.Organization.Equal(state.Organization) {
		updateParameters["org"] = plan.Organization.ValueString()
	}
	if !plan.StreetAddress.Equal(state.StreetAddress) {
		updateParameters["street"] = plan.StreetAddress.ValueString()
	}
	if !plan.City.Equal(state.City) {
		updateParameters["city"] = plan.City.ValueString()
	}
	if !plan.PostalCode.Equal(state.PostalCode) {
		updateParameters["pc"] = plan.PostalCode.ValueString()
	}
	if !plan.StateProvince.Equal(state.StateProvince) {
		updateParameters["sp"] = plan.StateProvince.ValueString()
	}
	if !plan.CountryCode.Equal(state.CountryCode) {
		updateParameters["cc"] = plan.CountryCode.ValueString()
	}
	if !plan.PhoneNumber.Equal(state.PhoneNumber) {
		updateParameters["voice"] = plan.PhoneNumber.ValueString()
	}
	if !plan.Fax.Equal(state.Fax) {
		updateParameters["fax"] = plan.Fax.ValueString()
	}
	if !plan.Email.Equal(state.Email) {
		updateParameters["email"] = plan.Email.ValueString()
	}
	if !plan.Remarks.Equal(state.Remarks) {
		updateParameters["remarks"] = plan.Remarks.ValueString()
	}
	if !plan.WhoisProtection.Equal(state.WhoisProtection) {
		updateParameters["protection"] = plan.WhoisProtection.ValueBool()
	}

	// Call API to update the contact
	apiResponse, err := r.client.Call(ctx, "contact.update", updateParameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update Contact",
			err.Error(),
		)
		return
	}

	if apiResponse.Code() != api.COMMAND_SUCCESSFUL && apiResponse.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError(
			"Failed to Update Contact",
			fmt.Sprintf("API response: %s", apiResponse.ApiError()),
		)
		return
	}

	// Set updated state
	plan.ID = state.ID // Ensure ID is retained
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *domainContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainContactModel

	// Retrieve the current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	contactID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Contact ID",
			fmt.Sprintf("Failed to parse Contact ID: %s", err.Error()),
		)
		return
	}

	// Prepare delete parameters
	deleteParameters := map[string]interface{}{
		"id": contactID,
	}

	// Call API to delete the contact
	apiResponse, err := r.client.Call(ctx, "contact.delete", deleteParameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete Contact",
			err.Error(),
		)
		return
	}

	if apiResponse.Code() != api.COMMAND_SUCCESSFUL && apiResponse.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError(
			"Failed to Delete Contact",
			fmt.Sprintf("API response: %s", apiResponse.ApiError()),
		)
		return
	}

	// Resource successfully deleted
	resp.State.RemoveResource(ctx)
}

// resolveID resolves the contact ID from API response data
func resolveID(rawID interface{}) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch id := rawID.(type) {
	case string:
		// When the contact already exists: id is a string
		return id, diags
	case float64:
		// When the contact is newly created: id is a float64
		return strconv.Itoa(int(id)), diags
	default:
		// Unknown type, return an error
		diags.AddError(
			"Unknown Type for Contact ID",
			fmt.Sprintf("API returned unexpected type for contact ID: %s", reflect.TypeOf(rawID)),
		)
		return "", diags
	}
}

func expandContactFromInfoResponse(contactData map[string]interface{}) (*domainContactModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Extract values from the contactData map
	id, ok := contactData["id"].(string)
	if !ok {
		diags.AddError(
			"Invalid Contact ID",
			"Expected 'id' to be a string but got a different type.",
		)
		return nil, diags
	}

	name, ok := contactData["name"].(string)
	if !ok {
		diags.AddError(
			"Invalid Name",
			"Expected 'name' to be a string but got a different type.",
		)
		return nil, diags
	}

	// Parse additional fields with safe type assertions
	organization := contactData["org"].(string)
	streetAddress := contactData["street"].(string)
	city := contactData["city"].(string)
	postalCode := contactData["pc"].(string)
	stateProvince := contactData["sp"].(string)
	countryCode := contactData["cc"].(string)
	phoneNumber := contactData["voice"].(string)
	faxNumber := contactData["fax"].(string)
	email := contactData["email"].(string)
	remarks := contactData["remarks"].(string)
	whoisProtection, err := strconv.ParseBool(contactData["protection"].(string))
	if err != nil {
		diags.AddError(
			"Invalid Whois Protection",
			"Expected 'protection' to be a boolean in string format but conversion failed.",
		)
		return nil, diags
	}

	// Create the ContactModel
	contact := &domainContactModel{
		ID:              types.StringValue(id),
		Type:            types.StringValue(contactData["type"].(string)),
		Name:            types.StringValue(name),
		Organization:    types.StringValue(organization),
		StreetAddress:   types.StringValue(streetAddress),
		City:            types.StringValue(city),
		PostalCode:      types.StringValue(postalCode),
		StateProvince:   types.StringValue(stateProvince),
		CountryCode:     types.StringValue(countryCode),
		PhoneNumber:     types.StringValue(phoneNumber),
		Fax:             types.StringValue(faxNumber),
		Email:           types.StringValue(email),
		Remarks:         types.StringValue(remarks),
		WhoisProtection: types.BoolValue(whoisProtection),
	}

	return contact, diags
}
