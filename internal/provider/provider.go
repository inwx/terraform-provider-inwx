package provider

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform/internal"
	"github.com/inwx/terraform/internal/api"
	"net/url"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": {
				Type:        schema.TypeString,
				Description: "",
				Optional:    true,
				Default:     "https://api.domrobot.com/jsonrpc/",
			},
			"username": {
				Type:        schema.TypeString,
				Description: "",
				Required:    true,
				Sensitive:   true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: "",
				Required:    true,
				Sensitive:   true,
			},
			"debug": {
				Type:        schema.TypeBool,
				Description: "",
				Optional:    true,
				Default:     false,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"inwx_domain":  internal.DomainResource(),
			"inwx_contact": internal.ContactResource(),
		},
		ConfigureContextFunc: configureContext,
	}
}

func configureContext(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	username := data.Get("username").(string)
	password := data.Get("password").(string)
	apiUrl, err := url.Parse(data.Get("api_url").(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not configure context",
			Detail:   fmt.Sprintf("Could not parse api_url: %w", err),
		})
		return nil, diags
	}
	logger := logr.Discard()
	debug := data.Get("debug").(bool)

	client, err := api.NewClient(username, password, apiUrl, &logger, debug)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not configure context",
			Detail:   fmt.Sprintf("Could not create http client: %w", err),
		})
		return nil, diags
	}

	loginParams := map[string]interface{}{
		"user": username,
		"pass": password,
	}
	call, err := client.Call(ctx, "account.login", loginParams)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not configure context",
			Detail:   fmt.Sprintf("Could not authenticate at api via account.login: %w", err),
		})
		return nil, diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not configure context",
			Detail:   fmt.Sprintf("Could not authenticate at api via account.login. Got response: %s", call.ApiError()),
		})
		return nil, diags
	}

	return client, diags
}
