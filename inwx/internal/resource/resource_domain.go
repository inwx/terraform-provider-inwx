package resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform/inwx/internal/api"
	"strings"
)

func DomainResource() *schema.Resource {
	validRenewalModes := []string{
		"AUTORENEW",
		"AUTODELETE",
		"AUTOEXPIRE",
	}

	return &schema.Resource{
		CreateContext: resourceDomainCreate,
		ReadContext:   resourceDomainRead,
		UpdateContext: resourceDomainUpdate,
		DeleteContext: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
				data.Set("name", data.Id())
				return schema.ImportStatePassthroughContext(ctx, data, i)
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the domain",
				Required:    true,
			},
			"nameservers": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					MinItems: 1,
				},
				Optional:    true,
				Description: "Set of nameservers of the domain",
			},
			"period": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Registration period of the domain",
			},
			"renewal_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "AUTORENEW",
				Description: "Renewal mode of the domain. One of: " + strings.Join(validRenewalModes, ", "),
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					var diags diag.Diagnostics
					for _, validRenewalMode := range validRenewalModes {
						if validRenewalMode == i.(string) {
							return diags
						}
					}

					diags = append(diags, diag.Diagnostic{
						Severity:      diag.Error,
						Summary:       "Invalid contact type",
						Detail:        "Must be one of: " + strings.Join(validRenewalModes, ", "),
						AttributePath: path,
					})
					return diags
				},
			},
			"transfer_lock": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the domain transfer lock should be enabled",
			},
			"contacts": {
				Type:        schema.TypeSet,
				Required:    true,
				MaxItems:    1,
				MinItems:    1,
				Elem:        contactsSchemaResource(),
				Description: "Contacts of the domain",
			},
			"extra_data": {
				Type:        schema.TypeMap,
				Optional:    true,
				Default:     map[string]string{},
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Extra data, needed for some jurisdictions",
			},
		},
	}
}

func contactsSchemaResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"registrant": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Id of the registrant contact",
			},
			"admin": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Id of the admin contact",
			},
			"tech": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Id of the tech contact",
			},
			"billing": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Id of the billing contact",
			},
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	// map value interface{} is actually an 'int' value, but we cannot parse it correctly here
	contactIds := d.Get("contacts").(*schema.Set).List()[0].(map[string]interface{})

	parameters := map[string]interface{}{
		"domain":       d.Get("name").(string),
		"ns":           d.Get("nameservers").(*schema.Set).List(),
		"period":       d.Get("period").(string),
		"registrant":   contactIds["registrant"],
		"admin":        contactIds["admin"],
		"tech":         contactIds["tech"],
		"billing":      contactIds["billing"],
		"transferLock": d.Get("transfer_lock").(bool),
		"renewalMode":  d.Get("renewal_mode").(string),
	}
	if extraData, ok := d.GetOk("extra_data"); ok {
		parameters["extData"] = extraData
	}

	call, err := client.Call(ctx, "domain.create", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not create domain",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not create domain",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	d.SetId(d.Get("name").(string))

	return diags
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*api.Client)

	parameters := map[string]interface{}{
		"domain": d.Id(),
		"wide":   2,
	}

	call, err := client.Call(ctx, "domain.info", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get domain info",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get domain info",
			Detail:   fmt.Sprintf("API response not status code 1000. Got response: %s", call.ApiError()),
		})
		return diags
	}

	resData := call["resData"].(map[string]interface{})
	d.Set("name", resData["domain"])
	d.Set("nameservers", resData["ns"])
	d.Set("period", resData["period"])
	d.Set("renewal_mode", resData["renewalMode"])
	d.Set("transfer_lock", resData["transferLock"] == 1.0) // convert 1.0 to true. Must be a float!

	contacts := map[string]interface{}{}
	contacts["registrant"] = int(resData["registrant"].(float64))
	contacts["admin"] = int(resData["admin"].(float64))
	contacts["tech"] = int(resData["tech"].(float64))
	contacts["billing"] = int(resData["billing"].(float64))

	d.Set("contacts", schema.NewSet(schema.HashResource(contactsSchemaResource()), []interface{}{contacts}))
	d.Set("extra_data", resData["extData"])

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*api.Client)

	if d.HasChange("name") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "domain 'name' cannot be updated",
		})
		return diags
	}

	parameters := map[string]interface{}{
		"domain": d.Get("name"),
	}

	if d.HasChange("nameservers") {
		parameters["ns"] = d.Get("nameservers")
	}
	if d.HasChange("period") {
		parameters["period"] = d.Get("period")
	}
	if d.HasChange("renewal_mode") {
		parameters["renewalMode"] = d.Get("renewal_mode")
	}
	if d.HasChange("transfer_lock") {
		parameters["transferLock"] = d.Get("transfer_lock")
	}
	if d.HasChange("contacts") {
		contacts := d.Get("contacts").(*schema.Set).List()[0].(map[string]interface{})
		parameters["registrant"] = contacts["registrant"]
		parameters["admin"] = contacts["admin"]
		parameters["tech"] = contacts["tech"]
		parameters["billing"] = contacts["billing"]
	}
	if d.HasChange("extra_data") {
		parameters["extData"] = d.Get("extra_data")
	}

	call, err := client.Call(ctx, "domain.update", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get domain info",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get domain info",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	return diags
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*api.Client)

	parameters := map[string]interface{}{
		"domain": d.Get("name"),
	}

	call, err := client.Call(ctx, "domain.delete", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete domain",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete domain",
			Detail:   fmt.Sprintf("API response not status code 1000 pr 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	return diags
}

func validateCountryCode(i interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	countryCode := i.(string)
	if len(countryCode) != 2 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not validate country code",
			Detail: fmt.Sprintf("Expected a two digit country code, got '%s' "+
				"with (%d) digits", countryCode, len(countryCode)),
			AttributePath: path,
		})
		return diags
	}
	return diags
}
