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

func resourceNameserverRecordParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected attribute1:attribute2", id)
	}

	return parts[0], parts[1], nil
}

func NameserverRecordResource() *schema.Resource {
	validRecordTypes := []string{
		"A", "AAAA", "AFSDB", "ALIAS", "CAA", "CERT", "CNAME", "HINFO", "KEY", "LOC", "MX", "NAPTR", "NS", "OPENPGPKEY",
		"PTR", "RP", "SMIMEA", "SOA", "SRV", "SSHFP", "TLSA", "TXT", "URI", "URL",
	}

	validUrlRedirectTypes := []string{
		"HEADER301", "HEADER302", "FRAME",
	}

	return &schema.Resource{
		CreateContext: resourceNameserverRecordCreate,
		ReadContext:   resourceNameserverRecordRead,
		UpdateContext: resourceNameserverRecordUpdate,
		DeleteContext: resourceNameserverRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
				domain, id, err := resourceNameserverRecordParseId(d.Id())

				if err != nil {
					return nil, err
				}

				d.Set("domain", domain)
				d.SetId(fmt.Sprintf("%s:%s", domain, id))

				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Description: "Domain name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ro_id": {
				Description: "DNS domain id",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"type": {
				Description: "Type of the nameserver record. One of: " + strings.Join(validRecordTypes, ", "),
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					var diags diag.Diagnostics
					for _, validRecordType := range validRecordTypes {
						if validRecordType == i.(string) {
							return diags
						}
					}

					diags = append(diags, diag.Diagnostic{
						Severity:      diag.Error,
						Summary:       "Invalid type type",
						Detail:        "Must be one of: " + strings.Join(validRecordTypes, ", "),
						AttributePath: path,
					})
					return diags
				},
			},
			"content": {
				Description: "Content of the nameserver record",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name of the nameserver record",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"ttl": {
				Description: "TTL (time to live) of the nameserver record",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3600,
			},
			"prio": {
				Description: "Priority of the nameserver record",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
			},
			"url_redirect_type": {
				Description: "Type of the url redirection. One of: " + strings.Join(validUrlRedirectTypes, ", "),
				Type:        schema.TypeString,
				Optional:    true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					var diags diag.Diagnostics
					for _, validUrlRedirectType := range validUrlRedirectTypes {
						if validUrlRedirectType == i.(string) {
							return diags
						}
					}

					diags = append(diags, diag.Diagnostic{
						Severity:      diag.Error,
						Summary:       "Invalid url_redirect_type",
						Detail:        "Must be one of: " + strings.Join(validUrlRedirectTypes, ", "),
						AttributePath: path,
					})
					return diags
				},
			},
			"url_redirect_title": {
				Description: "Title of the frame redirection",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"url_redirect_description": {
				Description: "Description of the frame redirection",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"url_redirect_fav_icon": {
				Description: "FavIcon of the frame redirection",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"url_redirect_keywords": {
				Description: "Keywords of the frame redirection",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"url_append": {
				Description: "Append the path for redirection",
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
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

func resourceNameserverRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	domain := d.Get("domain").(string)

	parameters := map[string]interface{}{
		"domain":  domain,
		"type":    d.Get("type").(string),
		"content": d.Get("content").(string),
	}

	if roId, ok := d.GetOk("ro_id"); ok {
		parameters["roId"] = roId
	}
	if name, ok := d.GetOk("name"); ok {
		parameters["name"] = name
	}
	if ttl, ok := d.GetOk("ttl"); ok {
		parameters["ttl"] = ttl
	}
	if prio, ok := d.GetOk("prio"); ok {
		parameters["prio"] = prio
	}
	if urlRedirectType, ok := d.GetOk("url_redirect_type"); ok {
		parameters["urlRedirectType"] = urlRedirectType
	}
	if urlRedirectTitle, ok := d.GetOk("url_redirect_title"); ok {
		parameters["urlRedirectTitle"] = urlRedirectTitle
	}
	if urlRedirectDescription, ok := d.GetOk("url_redirect_description"); ok {
		parameters["urlRedirectDescription"] = urlRedirectDescription
	}
	if urlRedirectFavIcon, ok := d.GetOk("url_redirect_fav_icon"); ok {
		parameters["urlRedirectFavIcon"] = urlRedirectFavIcon
	}
	if urlRedirectKeywords, ok := d.GetOk("url_redirect_keywords"); ok {
		parameters["urlRedirectKeywords"] = urlRedirectKeywords
	}
	if urlAppend, ok := d.GetOk("url_append"); ok {
		parameters["urlAppend"] = urlAppend
	}
	if testing, ok := d.GetOk("testing"); ok {
		parameters["testing"] = testing
	}

	call, err := client.Call(ctx, "nameserver.createRecord", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not add nameserver record",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not add nameserver record",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	resData := call["resData"].(map[string]any)

	d.SetId(domain + ":" + strconv.Itoa(int(resData["id"].(float64))))

	resourceNameserverRecordRead(ctx, d, m)

	return diags
}

func resourceNameserverRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"domain": d.Get("domain"),
	}

	call, err := client.Call(ctx, "nameserver.info", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get nameserver info",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get nameserver info",
			Detail:   fmt.Sprintf("API response not status code 1000. Got response: %s", call.ApiError()),
		})
		return diags
	}

	records := call["resData"].(map[string]any)["record"].([]any)

	for _, record := range records {
		recordt := record.(map[string]any)

		if d.Get("domain").(string)+":"+strconv.Itoa(int(recordt["id"].(float64))) == d.Id() {
			d.Set("domain", d.Get("domain").(string))
			d.Set("type", recordt["type"].(string))
			d.Set("content", recordt["content"].(string))

			if val, ok := recordt["name"]; ok {
				d.Set("name", val.(string))
			}
			if val, ok := recordt["urlRedirectType"]; ok {
				d.Set("url_redirect_type", val.(string))
			}
			if val, ok := recordt["urlRedirectType"]; ok {
				d.Set("url_redirect_type", val.(string))
			}
			if val, ok := recordt["urlRedirectTitle"]; ok {
				d.Set("url_redirect_title", val.(string))
			}
			if val, ok := recordt["urlRedirectDescription"]; ok {
				d.Set("url_redirect_description", val.(string))
			}
			if val, ok := recordt["urlRedirectKeywords"]; ok {
				d.Set("url_redirect_keywords", val.(string))
			}
			if val, ok := recordt["urlRedirectFavIcon"]; ok {
				d.Set("url_redirect_fav_icon", val.(string))
			}
			if val, ok := recordt["urlAppend"]; ok {
				d.Set("url_append", val.(bool))
			}
			if val, ok := recordt["testing"]; ok {
				d.Set("testing", val.(bool))
			}
			if val, ok := recordt["ttl"]; ok {
				d.Set("ttl", val.(float64))
			}
			if val, ok := recordt["prio"]; ok {
				d.Set("prio", val.(float64))
			}
		}
	}

	return diags
}

func resourceNameserverRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		"id": id,
	}

	if d.HasChange("type") {
		parameters["type"] = d.Get("type").(string)
	}
	if d.HasChange("content") {
		parameters["content"] = d.Get("content").(string)
	}

	if name, ok := d.GetOk("name"); ok && d.HasChange("name") {
		parameters["name"] = name
	}
	if ttl, ok := d.GetOk("ttl"); ok && d.HasChange("ttl") {
		parameters["ttl"] = ttl
	}
	if prio, ok := d.GetOk("prio"); ok && d.HasChange("prio") {
		parameters["prio"] = prio
	}
	if urlRedirectType, ok := d.GetOk("url_redirect_type"); ok && d.HasChange("url_redirect_type") {
		parameters["urlRedirectType"] = urlRedirectType
	}
	if urlRedirectTitle, ok := d.GetOk("url_redirect_title"); ok && d.HasChange("url_redirect_title") {
		parameters["urlRedirectTitle"] = urlRedirectTitle
	}
	if urlRedirectDescription, ok := d.GetOk("url_redirect_description"); ok && d.HasChange("url_redirect_description") {
		parameters["urlRedirectDescription"] = urlRedirectDescription
	}
	if urlRedirectFavIcon, ok := d.GetOk("url_redirect_fav_icon"); ok && d.HasChange("url_redirect_fav_icon") {
		parameters["urlRedirectFavIcon"] = urlRedirectFavIcon
	}
	if urlRedirectKeywords, ok := d.GetOk("url_redirect_keywords"); ok && d.HasChange("url_redirect_keywords") {
		parameters["urlRedirectKeywords"] = urlRedirectKeywords
	}
	if urlAppend, ok := d.GetOk("url_append"); ok && d.HasChange("url_append") {
		parameters["urlAppend"] = urlAppend
	}
	if testing, ok := d.GetOk("testing"); ok && d.HasChange("testing") {
		parameters["testing"] = testing
	}

	call, err := client.Call(ctx, "nameserver.updateRecord", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not update nameserver record",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not update nameserver record",
			Detail:   fmt.Sprintf("API response not status code 1000. Got response: %s", call.ApiError()),
		})
		return diags
	}

	return diags
}

func resourceNameserverRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		"id": id,
	}

	if testing, ok := d.GetOk("testing"); ok {
		parameters["testing"] = testing
	}

	call, err := client.Call(ctx, "nameserver.deleteRecord", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete nameserver record",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete nameserver record",
			Detail:   fmt.Sprintf("API response not status code 1000 pr 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	return diags
}
