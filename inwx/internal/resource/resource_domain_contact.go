package resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/api"
	"strconv"
	"strings"
)

type Contact struct {
	Type            string
	Name            string
	Organization    string
	StreetAddress   string
	City            string
	PostalCode      string
	StateProvince   string
	CountryCode     string
	PhoneNumber     string
	FaxNumber       string
	Email           string
	Remarks         string
	WhoisProtection bool
}

func DomainContactResource() *schema.Resource {
	validContactTypes := []string{
		"ORG",
		"PERSON",
		"ROLE",
	}

	return &schema.Resource{
		CreateContext: resourceContactCreate,
		ReadContext:   resourceContactRead,
		UpdateContext: resourceContactUpdate,
		DeleteContext: resourceContactDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of contact. One of: " + strings.Join(validContactTypes, ", "),
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					var diags diag.Diagnostics
					for _, validContactType := range validContactTypes {
						if validContactType == i.(string) {
							return diags
						}
					}

					diags = append(diags, diag.Diagnostic{
						Severity:      diag.Error,
						Summary:       "Invalid contact type",
						Detail:        "Must be one of: " + strings.Join(validContactTypes, ", "),
						AttributePath: path,
					})
					return diags
				},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "First and lastname of the contact",
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The legal name of the organization. Required for types other than person",
			},
			"street_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Street Address of the contact",
			},
			"city": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "City of the contact",
			},
			"postal_code": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Postal Code/Zipcode of the contact",
			},
			"state_province": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "State/Province name of the contact",
			},
			"country_code": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateCountryCode,
				Description:      "Country code of the contact. Must be two characters",
			},
			"phone_number": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Phone number of the contact",
			},
			"fax": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Fax number of the contact",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Contact email address",
			},
			"remarks": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Custom description of the contact",
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					var diags diag.Diagnostics
					remarks := i.(string)
					if len(remarks) > 255 {
						diags = append(diags, diag.Diagnostic{
							Severity:      diag.Error,
							Summary:       "Remarks is too long",
							Detail:        "Maximum allowed length is 255 characters",
							AttributePath: path,
						})
					}
					return diags
				},
			},
			"whois_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Whether whois protection for the contact should be enabled. " +
					"Depends on the registry supporting it. Not the same as whois protection for a domain",
			},
		},
	}
}

func resourceContactCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*api.Client)

	contact := expandContactFromResourceData(data)

	parameters := map[string]interface{}{
		"type":       contact.Type,
		"name":       contact.Name,
		"street":     contact.StreetAddress,
		"city":       contact.City,
		"pc":         contact.PostalCode,
		"sp":         contact.StateProvince,
		"cc":         contact.CountryCode,
		"voice":      contact.PhoneNumber,
		"email":      contact.Email,
		"protection": contact.WhoisProtection,
	}
	if contact.Organization != "" {
		parameters["org"] = contact.Organization
	}
	if contact.StateProvince != "" {
		parameters["sp"] = contact.StateProvince
	}
	if contact.FaxNumber != "" {
		parameters["fax"] = contact.FaxNumber
	}
	if contact.Remarks != "" {
		parameters["remarks"] = contact.Remarks
	}

	call, err := client.Call(ctx, "contact.create", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not create contact",
			Detail:   err.Error(),
		})
		return diags
	}

	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not create contact",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	data.SetId(strconv.Itoa(int(call["resData"].(map[string]interface{})["id"].(float64))))
	return diags
}

func resourceContactRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*api.Client)

	contactId, err := strconv.Atoi(data.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not read numerical contact id",
			Detail:   err.Error(),
		})
		return diags
	}
	parameters := map[string]interface{}{
		"id":   contactId,
		"wide": 2,
	}

	call, err := client.Call(ctx, "contact.info", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get contact info",
			Detail:   err.Error(),
		})
		return diags
	}

	contact := expandContactFromInfoResponse(call["resData"].(map[string]interface{})["contact"].(map[string]interface{}))

	data.Set("type", contact.Type)
	data.Set("name", contact.Name)
	if contact.Organization != "" {
		data.Set("organization", contact.Organization)
	}
	data.Set("street_address", contact.StreetAddress)
	data.Set("city", contact.City)
	data.Set("postal_code", contact.PostalCode)
	if contact.StateProvince != "" {
		data.Set("state_province", contact.StateProvince)
	}
	data.Set("country_code", contact.CountryCode)
	data.Set("phone_number", contact.PhoneNumber)
	if contact.FaxNumber != "" {
		data.Set("fax", contact.FaxNumber)
	}
	data.Set("email", contact.Email)
	if contact.Remarks != "" {
		data.Set("remarks", contact.Remarks)
	}

	return diags
}

func resourceContactUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*api.Client)

	parameters := map[string]interface{}{
		"id": data.Id(),
	}

	if data.HasChange("type") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "contact 'type' cannot be updated.",
		})
		return diags
	}
	if data.HasChange("name") {
		parameters["name"] = data.Get("name")
	}
	if data.HasChange("organization") {
		parameters["org"] = data.Get("organization")
	}
	if data.HasChange("street_address") {
		parameters["street"] = data.Get("street_address")
	}
	if data.HasChange("city") {
		parameters["city"] = data.Get("city")
	}
	if data.HasChange("postal_code") {
		parameters["pc"] = data.Get("postal_code")
	}
	if data.HasChange("state_province") {
		parameters["sp"] = data.Get("state_province")
	}
	if data.HasChange("country_code") {
		parameters["cc"] = data.Get("country_code")
	}
	if data.HasChange("phone_number") {
		parameters["voice"] = data.Get("phone_number")
	}
	if data.HasChange("fax") {
		parameters["fax"] = data.Get("fax")
	}
	if data.HasChange("email") {
		parameters["email"] = data.Get("email")
	}
	if data.HasChange("remarks") {
		parameters["remarks"] = data.Get("remarks")
	}
	if data.HasChange("whois_protection") {
		parameters["protection"] = data.Get("whois_protection")
	}

	call, err := client.Call(ctx, "contact.update", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not update contact",
			Detail:   err.Error(),
		})
		return diags
	}

	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not update contact",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	return diags
}

func resourceContactDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*api.Client)

	contactId, err := strconv.Atoi(data.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not read numerical contact id",
			Detail:   err.Error(),
		})
		return diags
	}
	parameters := map[string]interface{}{
		"id": contactId,
	}
	call, err := client.Call(ctx, "contact.delete", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete contact",
			Detail:   err.Error(),
		})
		return diags
	}

	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete contact",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	return diags
}

func expandContactFromResourceData(data *schema.ResourceData) *Contact {
	var organization string
	if dataOrganization, ok := data.GetOk("organization"); ok {
		organization = dataOrganization.(string)
	}
	var stateProvince string
	if dataStateProvince, ok := data.GetOk("state_province"); ok {
		stateProvince = dataStateProvince.(string)
	}
	var fax string
	if dataFax, ok := data.GetOk("fax"); ok {
		fax = dataFax.(string)
	}
	var remarks string
	if dataRemarks, ok := data.GetOk("remarks"); ok {
		remarks = dataRemarks.(string)
	}

	return &Contact{
		Type:            data.Get("type").(string),
		Name:            data.Get("name").(string),
		Organization:    organization,
		StreetAddress:   data.Get("street_address").(string),
		City:            data.Get("city").(string),
		PostalCode:      data.Get("postal_code").(string),
		StateProvince:   stateProvince,
		CountryCode:     data.Get("country_code").(string),
		PhoneNumber:     data.Get("phone_number").(string),
		FaxNumber:       fax,
		Email:           data.Get("email").(string),
		Remarks:         remarks,
		WhoisProtection: data.Get("whois_protection").(bool),
	}
}

func expandContactFromInfoResponse(contactData map[string]interface{}) *Contact {
	var organization string
	if dataOrganization, ok := contactData["org"]; ok {
		organization = dataOrganization.(string)
	}
	var stateProvince string
	if dataStateProvince, ok := contactData["sp"]; ok {
		stateProvince = dataStateProvince.(string)
	}
	var fax string
	if dataFax, ok := contactData["fax"]; ok {
		fax = dataFax.(string)
	}
	var remarks string
	if dataRemarks, ok := contactData["remarks"]; ok {
		remarks = dataRemarks.(string)
	}

	// why is this a boolean in a string ffs
	whoisProtection, err := strconv.ParseBool(contactData["protection"].(string))
	if err != nil {
		panic("api error. expected 'protection' boolean string could not be converted to actual boolean value")
	}

	return &Contact{
		Type:            contactData["type"].(string),
		Name:            contactData["name"].(string),
		Organization:    organization,
		StreetAddress:   contactData["street"].(string),
		City:            contactData["city"].(string),
		PostalCode:      contactData["pc"].(string),
		StateProvince:   stateProvince,
		CountryCode:     contactData["cc"].(string),
		PhoneNumber:     contactData["voice"].(string),
		FaxNumber:       fax,
		Email:           contactData["email"].(string),
		Remarks:         remarks,
		WhoisProtection: whoisProtection,
	}
}
