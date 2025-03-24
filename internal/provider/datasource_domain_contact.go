package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/inwx/terraform-provider-inwx/internal/api"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type domainContactDataSource struct {
	client *api.Client
}

func (d *domainContactDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "inwx_domain_contact"
}

func NewDomainContactDataSource() datasource.DataSource {
	return &domainContactDataSource{}
}

func (d *domainContactDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				Computed:    true,
			},
		},
	}
}

func (d *domainContactDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data domainContactModel
	diags := req.Config.Get(ctx, &data)
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
	call, err := d.client.Call(ctx, "contact.info", parameters)
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
