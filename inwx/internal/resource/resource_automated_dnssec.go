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
	}
}

func resourceAutomatedDNSSECRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"domains": []string{d.Get("domain").(string)},
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

	records := call["resData"].(map[string]any)["record"].([]any)

	for _, record := range records {
		recordt := record.(map[string]any)

		if recordt["domain"].(string) == d.Get("domain").(string) && recordt["dnssecStatus"].(string) == "AUTO" {
			d.SetId(recordt["domain"].(string))
		}
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
			Detail:   fmt.Sprintf("API response not status code 1000 pr 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	return diags
}
