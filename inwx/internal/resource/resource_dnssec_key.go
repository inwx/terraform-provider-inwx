package resource

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/api"
)

func DNSSECKeyResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDNSSECKeyCreate,
		ReadContext:   resourceDNSSECKeyRead,
		DeleteContext: resourceDNSSECKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 2 {
					return nil, errors.New("invalid resource import specifier. Use: terraform import <domain>/<digest>")
				}

				_ = d.Set("domain", parts[0])
				_ = d.Set("digest", parts[1])

				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Description: "Name of the domain",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"public_key": {
				Description: "Public key of the domain",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"algorithm": {
				Description: "Algorithm used for the public key",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"digest": {
				Description: "Computed digest for the public key",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"digest_type": {
				Description: "Digest type",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"flag": {
				Description: "Key flag (256=ZSK, 257=KSK)",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"key_tag": {
				Description: "Key tag",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"status": {
				Description: "DNSSEC status",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceDNSSECKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"domainName": d.Get("domain").(string),
		"dnskey": fmt.Sprintf(
			"%s. IN DNSKEY 257 3 %d %s",
			d.Get("domain").(string),
			d.Get("algorithm").(int),
			d.Get("public_key").(string),
		),
		"calculateDigest": true,
	}

	call, err := client.Call(ctx, "dnssec.adddnskey", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not add DNSKEY",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not add DNSKEY",
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	resData := call["resData"].(map[string]interface{})

	parts := strings.Split(resData["ds"].(string), " ")
	if len(parts) != 4 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not parse returned DS",
			Detail:   fmt.Sprintf("API response not in expected format. Got response: %s", resData["ds"]),
		})
		return diags
	}

	d.Set("digest", parts[3])

	resourceDNSSECKeyRead(ctx, d, m)

	return diags
}

func resourceDNSSECKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"domainName": d.Get("domain").(string),
		"digest":     d.Get("digest").(string),
		"active":     1,
	}

	call, err := client.Call(ctx, "dnssec.listkeys", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get DNSSEC keys",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get DNSSEC keys",
			Detail:   fmt.Sprintf("API response not status code 1000. Got response: %s", call.ApiError()),
		})
		return diags
	}

	// Safely handle missing or empty resData
	rawResData, exists := call["resData"]
	if !exists || rawResData == nil {
		// Mark resource as deleted (removes it from the Terraform state)
		d.SetId("")
		return diags
	}

	// Cast resData to expected type, with validation
	resData, ok := rawResData.([]interface{})
	if !ok || len(resData) == 0 {
		// Mark resource as deleted if resData is not valid or empty
		d.SetId("")
		return diags
	}

	// Proceed with normal processing
	key := resData[0].(map[string]interface{})

	d.SetId(key["id"].(string))
	d.Set("domain", key["ownerName"].(string))
	d.Set("public_key", key["publicKey"].(string))
	d.Set("digest", key["digest"].(string))
	d.Set("status", key["status"].(string))

	if i, err := strconv.Atoi(key["algorithmId"].(string)); err == nil {
		d.Set("algorithm", i)
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "algorithm: failed to parse int from string",
			Detail:   err.Error(),
		})
	}

	if i, err := strconv.Atoi(key["digestTypeId"].(string)); err == nil {
		d.Set("digest_type", i)
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "digest_type: failed to parse int from string",
			Detail:   err.Error(),
		})
	}

	if i, err := strconv.Atoi(key["flagId"].(string)); err == nil {
		d.Set("flag", i)
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "flag: failed to parse int from string",
			Detail:   err.Error(),
		})
	}

	if i, err := strconv.Atoi(key["keyTag"].(string)); err == nil {
		d.Set("key_tag", i)
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "key_tag: failed to parse int from string",
			Detail:   err.Error(),
		})
	}

	return diags
}

func resourceDNSSECKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"key": d.Id(),
	}

	call, err := client.Call(ctx, "dnssec.deletednskey", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete DNSKEY",
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete DNSKEY",
			Detail:   fmt.Sprintf("API response not status code 1000 pr 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	return diags
}
