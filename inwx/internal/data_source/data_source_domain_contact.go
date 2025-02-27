package data_source

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/resource"
	"strings"
)

func DomainContactDataSource() *schema.Resource {
	validContactTypes := []string{
		"ORG",
		"PERSON",
		"ROLE",
	}

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Type of contact. One of: " + strings.Join(validContactTypes, ", "),
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "First and lastname of the contact",
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The legal name of the organization. Required for types other than person",
			},
			"street_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Street Address of the contact",
			},
			"city": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "City of the contact",
			},
			"postal_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Postal Code/Zipcode of the contact",
			},
			"state_province": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "State/Province name of the contact",
			},
			"country_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Country code of the contact. Must be two characters",
			},
			"phone_number": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Phone number of the contact",
			},
			"fax": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Fax number of the contact",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Contact email address",
			},
			"remarks": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Custom description of the contact",
			},
		},
		ReadContext: resourceContactRead,
		Description: "Data source provides information about a domain contact",
	}
}

func resourceContactRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	parameters := map[string]interface{}{
		"id":   data.Get("id").(int),
		"wide": 2,
	}

	return resource.AbstractResourceContactRead(ctx, data, meta, parameters)
}
