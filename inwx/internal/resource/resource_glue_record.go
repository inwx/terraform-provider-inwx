package resource

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/api"
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
				hostname, id, err := resourceGlueRecordParseId(d.Id())
				if err != nil {
					return nil, err
				}

				d.Set("hostname", hostname)
				d.SetId(fmt.Sprintf("%s:%s", hostname, id))

				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"hostname": {
				Description: "Name of host",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ip": {
				Description: "Ip address(es)",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"status": {
				Description: "Status of the hostname",
				Type:        schema.TypeString,
				Computed:    true,
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
		"ip":       d.Get("ip").([]interface{}),
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
	roId := strconv.Itoa(int(resData["roId"].(float64)))

	d.SetId(fmt.Sprintf("%s:%s", hostname, roId))

	return resourceGlueRecordRead(ctx, d, m)
}

func resourceGlueRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	hostname, roId, err := resourceGlueRecordParseId(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not parse id",
			Detail:   err.Error(),
		})
		return diags
	}

	parameters := map[string]interface{}{
		"hostname": hostname,
		"roId":     roId,
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

	resData, ok := call["resData"].(map[string]any)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get glue record info",
			Detail:   "unexpected response format: missing resData",
		})
		return diags
	}

	// Check if the returned record matches our ID
	roIdFloat, ok := resData["roId"].(float64)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get glue record info",
			Detail:   "unexpected response format: missing or invalid roId",
		})
		return diags
	}

	if hostname+":"+strconv.Itoa(int(roIdFloat)) != d.Id() {
		d.SetId("") // Resource not found
		return diags
	}

	d.Set("hostname", resData["hostname"].(string))
	if status, ok := resData["status"].(string); ok {
		d.Set("status", status)
	}
	if ips, ok := resData["ip"].([]interface{}); ok {
		d.Set("ip", ips)
	}

	return diags
}

func resourceGlueRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	_, roId, err := resourceGlueRecordParseId(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not parse id",
			Detail:   err.Error(),
		})
		return diags
	}

	parameters := map[string]interface{}{
		"roId": roId,
		"ip":   d.Get("ip").([]interface{}),
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

	hostname, roId, err := resourceGlueRecordParseId(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not parse id",
			Detail:   err.Error(),
		})
		return diags
	}

	parameters := map[string]interface{}{
		"hostname": hostname,
		"roId":     roId,
	}

	if testing, ok := d.GetOk("testing"); ok {
		parameters["testing"] = testing
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
