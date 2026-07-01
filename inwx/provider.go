package inwx

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/inwx/terraform-provider-inwx/inwx/internal/data_source"

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
			"shared_secret": {
				Type: schema.TypeString,
				Description: "Base32-encoded TOTP shared secret for 2FA. " +
					"The provider computes a fresh TAN from this secret on every login, " +
					"avoiding the 30-second expiry race between plan and apply. " +
					"Can be passed as `INWX_SHARED_SECRET` env var.",
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("INWX_SHARED_SECRET", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"inwx_domain":            resource.DomainResource(),
			"inwx_domain_contact":    resource.DomainContactResource(),
			"inwx_dnssec_key":        resource.DNSSECKeyResource(),
			"inwx_nameserver_record": resource.NameserverRecordResource(),
			"inwx_automated_dnssec":  resource.AutomatedDNSSECResource(),
			"inwx_nameserver":        resource.NameserverResource(),
			"inwx_glue_record":       resource.GlueRecordResource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"inwx_domain_contact": data_source.DomainContactDataSource(),
		},
		ConfigureContextFunc: configureContext,
	}
}

// generateTOTP computes an RFC 6238 TOTP code (HMAC-SHA1, 30s step, 6 digits)
// from a base32-encoded shared secret.
func generateTOTP(secret string) (string, error) {
	secret = strings.ToUpper(strings.TrimRight(secret, "="))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("invalid shared_secret: %v", err)
	}

	counter := uint64(time.Now().Unix() / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	code := (uint32(h[offset])&0x7f)<<24 |
		uint32(h[offset+1])<<16 |
		uint32(h[offset+2])<<8 |
		uint32(h[offset+3])
	code = code % 1_000_000

	return fmt.Sprintf("%06d", code), nil
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
			Detail:   fmt.Sprintf("Could not parse api_url: %v", err),
		})
		return nil, diags
	}
	logger := logr.Discard()

	client, err := api.NewClient(username, password, apiUrl, &logger, false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not configure context",
			Detail:   fmt.Sprintf("Could not create http client: %v", err),
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
			Detail:   fmt.Sprintf("Could not authenticate at api via account.login: %v", err),
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

	if call.Code() == 2200 {
		var tan string
		if sharedSecret, ok := data.GetOk("shared_secret"); ok && sharedSecret.(string) != "" {
			tan, err = generateTOTP(sharedSecret.(string))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Could not generate TOTP",
					Detail:   fmt.Sprintf("Could not compute TAN from shared_secret: %v", err),
				})
				return nil, diags
			}
		} else if t, ok := data.GetOk("tan"); ok && t.(string) != "" {
			tan = t.(string)
		}

		if tan != "" {
			call, err = client.Call(ctx, "account.unlock", map[string]interface{}{
				"tan": tan,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Could not unlock account",
					Detail:   fmt.Sprintf("Could not authenticate at api via account.unlock: %v", err),
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
	}

	return client, diags
}
