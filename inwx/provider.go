package inwx

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/api"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/resource"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": {
				Type: schema.TypeString,
				Description: "URL of the RPC API endpoint. Use `https://api.domrobot.com/jsonrpc/` " +
					"for production and `https://api.ote.domrobot.com/jsonrpc/` for tests. " +
					"Can be passed as `INWX_API_URL` env var.",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INWX_API_URL", "https://api.domrobot.com/jsonrpc/"),
			},
			"username": {
				Type:        schema.TypeString,
				Description: "Login username of the api. Can be passed as `INWX_USERNAME` env var.",
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("INWX_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Description: "Login password of the api. Can be passed as `INWX_PASSWORD` env var.",
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("INWX_PASSWORD", nil),
			},
			"tan": {
				Type:        schema.TypeString,
				Description: "Mobile-TAN to unlock account. Can be passed as `INWX_TAN` env var.",
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("INWX_TAN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"inwx_domain":            resource.DomainResource(),
			"inwx_domain_contact":    resource.DomainContactResource(),
			"inwx_dnssec_key":        resource.DNSSECKeyResource(),
			"inwx_nameserver_record": resource.NameserverRecordResource(),
			"inwx_nameserver":        resource.NameserverResource(),
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

	client, err := api.NewClient(username, password, apiUrl, &logger, false)
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
			Detail: fmt.Sprintf("Could not authenticate at api via account.login. "+
				"Got response: %s", call.ApiError()),
		})
		return nil, diags
	}

	call, err = client.Call(ctx, "account.info", map[string]interface{}{})

	if tan, ok := data.GetOk("tan"); ok && call.Code() == 2200 && tan != "" {
		call, err := client.Call(ctx, "account.unlock", map[string]interface{}{
			"tan": tan,
		})
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Could not unlock account",
				Detail:   fmt.Sprintf("Could not authenticate at api via account.unlock: %w", err),
			})
			return nil, diags
		}
		if call.Code() != api.COMMAND_SUCCESSFUL {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Could not unlock account",
				Detail: fmt.Sprintf("Could not authenticate at api via account.unlock. "+
					"Got response: %s", call.ApiError()),
			})
			return nil, diags
		}
	}

	return client, diags
}
