package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/api"
)

func AutomatedDNSSECResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAutomatedDNSSECCreate,
		DeleteContext: resourceAutomatedDNSSECDelete,
		ReadContext:   resourceAutomatedDNSSECRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Description: "Name of the domain",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
		Description:        "Manages automated DNSSEC for a domain",
		DeprecationMessage: "",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceAutomatedDNSSECRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	domain := d.Get("domain").(string)
	if domain == "" {
		domain = d.Id()
		if err := d.Set("domain", domain); err != nil {
			return diag.FromErr(err)
		}
	}

	parameters := map[string]interface{}{
		"domains": []string{domain},
	}

	call, err := client.Call(ctx, "dnssec.info", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not read DNSSEC info",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not read DNSSEC info",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	resData, ok := call["resData"].(map[string]any)
	if !ok {
		d.SetId("") // Clear ID if resData is not found
		return diags
	}

	data, ok := resData["data"].([]any)
	if !ok || len(data) == 0 {
		d.SetId("") // Clear ID if no data found
		return diags
	}

	found := false
	for _, item := range data {
		domainInfo, ok := item.(map[string]any)
		if !ok {
			continue
		}

		if domainInfo["domain"].(string) == domain && domainInfo["dnssecStatus"].(string) == "AUTO" {
			d.SetId(domain)
			found = true
			break
		}
	}

	if !found {
		d.SetId("") // Clear ID if no matching record found
	}

	return diags
}

func resourceAutomatedDNSSECCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"domainName": d.Get("domain").(string),
	}

	call, err := client.Call(ctx, "dnssec.enablednssec", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not enable automated DNSSEC",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not enable automated DNSSEC",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	d.SetId(d.Get("domain").(string))

	return diags
}

func resourceAutomatedDNSSECDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"domainName": d.Get("domain").(string),
	}

	call, err := client.Call(ctx, "dnssec.disablednssec", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not disable automated DNSSEC",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not disable automated DNSSEC",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	return diags
}
