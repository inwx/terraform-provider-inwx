package resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/api"
	"strconv"
	"strings"
)

func resourceGlueRecordParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected attribute1:attribute2", id)
	}

	return parts[0], parts[1], nil
}

func GlueRecordResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGlueRecordCreate,
		ReadContext:   resourceGlueRecordRead,
		UpdateContext: resourceGlueRecordUpdate,
		DeleteContext: resourceGlueRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
				domain, id, err := resourceGlueRecordParseId(d.Id())

				if err != nil {
					return nil, err
				}

				d.Set("domain", domain)
				d.SetId(fmt.Sprintf("%s:%s", domain, id))

				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"hostname": {
				Description: "Name of host",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ro_id": {
				Description: "Id (Repository Object Identifier) of the hostname",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"ip": {
				Description: "Ip address(es)",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"testing": {
				Description: "Execute command in testing mode",
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
			},
		},
	}
}

func resourceGlueRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	hostname := d.Get("hostname").(string)

	parameters := map[string]interface{}{
		"hostname": hostname,
		"ip":       d.Get("ip").(string),
	}

	if testing, ok := d.GetOk("testing"); ok {
		parameters["testing"] = testing
	}

	call, err := client.Call(ctx, "host.create", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not create glue host",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not create glue host",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	resData := call["resData"].(map[string]any)

	d.SetId(hostname + ":" + strconv.Itoa(int(resData["roId"].(float64))))

	resourceNameserverRecordRead(ctx, d, m)

	return diags
}

func resourceGlueRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"hostname": d.Get("hostname"),
	}

	call, err := client.Call(ctx, "host.info", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get glue record info",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get glue record info",
			Detail:   fmt.Sprintf("API response not status code 1000. Got response: %s", call.ApiError()),
		})
		return diags
	}

	records := call["resData"].(map[string]any)["record"].([]any)

	for _, record := range records {
		recordt := record.(map[string]any)

		if d.Get("hostname").(string)+":"+strconv.Itoa(int(recordt["roId"].(float64))) == d.Id() {
			d.Set("ro_id", d.Get("ro_id").(string))
			d.Set("hostname", d.Get("hostname").(string))
			d.Set("status", recordt["status"].(string))
			d.Set("ip", recordt["ip"].(string))
		}
	}

	return diags
}

func resourceGlueRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	_, id, err := resourceGlueRecordParseId(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not parse id",
			Detail:   err.Error(),
		})
		return diags
	}

	parameters := map[string]interface{}{
		"roId": id,
		"ip":   d.Get("ip").(string),
	}

	if d.HasChange("hostname") {
		parameters["hostname"] = d.Get("hostname").(string)
	}
	if testing, ok := d.GetOk("testing"); ok && d.HasChange("testing") {
		parameters["testing"] = testing
	}

	err = client.CallNoResponseBody(ctx, "host.update", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not update glue record",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceGlueRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	_, id, err := resourceNameserverRecordParseId(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not parse id",
			Detail:   err.Error(),
		})
		return diags
	}

	parameters := map[string]interface{}{
		"roId": id,
	}

	if hostname, ok := d.GetOk("hostname"); ok {
		parameters["hostname"] = hostname
	}
	if testing, ok := d.GetOk("testing"); ok {
		parameters["roId"] = testing
	}

	err = client.CallNoResponseBody(ctx, "host.delete", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete glue record",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}
