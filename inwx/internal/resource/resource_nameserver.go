package resource

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/api"
)

func resourceNameserverParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected attribute1:attribute2", id)
	}

	return parts[0], parts[1], nil
}

func NameserverResource() *schema.Resource {
	validTypes := []string{
		"MASTER", "SLAVE",
	}

	validUrlRedirectTypes := []string{
		"HEADER301", "HEADER302", "FRAME",
	}

	return &schema.Resource{
		CreateContext: resourceNameserverCreate,
		ReadContext:   resourceNameserverRead,
		UpdateContext: resourceNameserverUpdate,
		DeleteContext: resourceNameserverDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
				domain, id, err := resourceNameserverParseId(d.Id())
				if err != nil {
					return nil, err
				}

				err = d.Set("domain", domain)
				if err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("%s:%s", domain, id))

				diags := resourceNameserverRead(ctx, d, i)
				if diags.HasError() {
					return nil, fmt.Errorf("failed to read nameserver data: %v", diags[0].Summary)
				}

				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Description: "Domain name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Description: "Type of the nameserver. One of: " + strings.Join(validTypes, ", "),
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					var diags diag.Diagnostics
					for _, validRecordType := range validTypes {
						if validRecordType == i.(string) {
							return diags
						}
					}

					diags = append(diags, diag.Diagnostic{
						Severity:      diag.Error,
						Summary:       "Invalid type type",
						Detail:        "Must be one of: " + strings.Join(validTypes, ", "),
						AttributePath: path,
					})
					return diags
				},
				ForceNew: false,
			},
			"nameservers": {
				Description: "List of nameservers",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: false,
			},
			"master_ip": {
				Description: "Master IP address",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
			},
			"web": {
				Description: "Web nameserver entry",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
			},
			"mail": {
				Description: "Mail nameserver entry",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
			},
			"soa_mail": {
				Description: "Email address for SOA record",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
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
				ForceNew: false,
			},
			"url_redirect_title": {
				Description: "Title of the frame redirection",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
			},
			"url_redirect_description": {
				Description: "Description of the frame redirection",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
			},
			"url_redirect_fav_icon": {
				Description: "FavIcon of the frame redirection",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
			},
			"url_redirect_keywords": {
				Description: "Keywords of the frame redirection",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
			},
			"testing": {
				Description: "Execute command in testing mode",
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				ForceNew:    false,
			},
			"ignore_existing": {
				Description: "Ignore existing",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

// buildNameserverParameters builds the common parameter map used by both create and update operations
func buildNameserverParameters(d *schema.ResourceData) map[string]interface{} {
	parameters := map[string]interface{}{
		"domain": d.Get("domain").(string),
		"type":   d.Get("type").(string),
	}

	// Add optional parameters if they exist
	if ns, ok := d.GetOk("nameservers"); ok {
		parameters["ns"] = ns
	}
	if masterIp, ok := d.GetOk("master_ip"); ok {
		parameters["masterIp"] = masterIp
	}
	if web, ok := d.GetOk("web"); ok {
		parameters["web"] = web
	}
	if mail, ok := d.GetOk("mail"); ok {
		parameters["mail"] = mail
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
	if testing, ok := d.GetOk("testing"); ok {
		parameters["testing"] = testing
	}

	return parameters
}

// createOrUpdateNameserver handles both create and update operations
func createOrUpdateNameserver(ctx context.Context, d *schema.ResourceData, m interface{}, isCreate bool) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := buildNameserverParameters(d)

	// Add create-specific parameters
	if isCreate {
		if ignoreExisting, ok := d.GetOk("ignore_existing"); ok {
			parameters["ignoreExisting"] = ignoreExisting
		}
	}

	// Determine the API endpoint
	endpoint := "nameserver.update"
	if isCreate {
		endpoint = "nameserver.create"
	}

	call, err := client.Call(ctx, endpoint, parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Could not %s nameserver record", map[bool]string{true: "add", false: "update"}[isCreate]),
			Detail:   err.Error(),
		})
		return diags
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Could not %s nameserver record", map[bool]string{true: "add", false: "update"}[isCreate]),
			Detail:   fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		})
		return diags
	}

	// Set the ID for newly created resources
	if isCreate {
		resData := call["resData"].(map[string]any)
		d.SetId(d.Get("domain").(string) + ":" + strconv.Itoa(int(resData["roId"].(float64))))
	}

	// Read the resource to ensure the Terraform state is up to date
	return resourceNameserverRead(ctx, d, m)
}

func resourceNameserverCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createOrUpdateNameserver(ctx, d, m, true)
}

func resourceNameserverRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"domain": d.Get("domain"),
	}

	call, err := client.Call(ctx, "nameserver.info", parameters)
	if err != nil {
		return diags
	}

	if resData, ok := call["resData"]; ok {
		resData := resData.(map[string]any)

		// Helper function to safely set values
		setValue := func(key string, field string) {
			if val, ok := resData[key]; ok {
				err := d.Set(field, val)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("Could not set %s", field),
						Detail:   fmt.Sprintf("Expected %s. %s", field, err.Error()),
					})
				}
			}
		}

		var nsInterface []interface{}
		// Check if "record" key exists and is an array
		if records, ok := resData["record"].([]interface{}); ok {
			for _, item := range records {
				// Convert item to a map
				if recordMap, ok := item.(map[string]interface{}); ok {
					// Check if "type" is "NS" and get "content"
					if recordType, ok := recordMap["type"].(string); ok && recordType == "NS" {
						if content, ok := recordMap["content"].(string); ok {
							nsInterface = append(nsInterface, content)
						}
					}
					// Check if "type" is "SOA" and get "content"
					if recordType, ok := recordMap["type"].(string); ok && recordType == "SOA" {
						if content, ok := recordMap["content"].(string); ok {
							// soa-content split
							parts := strings.Fields(content)
							if len(parts) > 1 {
								// second part is rname
								rname := parts[1]

								soaMail := transformRname(rname)

								if err := d.Set("soa_mail", soaMail); err != nil {
									diags = append(diags, diag.Diagnostic{
										Severity: diag.Error,
										Summary:  fmt.Sprintf("Could not set %s", rname),
										Detail:   fmt.Sprintf("Expected %s. %s", rname, err.Error()),
									})
								}
							}
						}
					}
				}
			}
		}
		if err := d.Set("nameservers", nsInterface); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Could not set nameservers %s", err.Error()),
				Detail:   fmt.Sprintf("Expected %s. %s", nsInterface, err.Error()),
			})
		}

		setValue("domain", "domain")
		setValue("type", "type")
		setValue("masterIp", "master_ip")
	}

	return diags
}

func resourceNameserverDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*api.Client)

	parameters := map[string]interface{}{
		"domain": d.Get("domain"),
	}

	if testing, ok := d.GetOk("testing"); ok {
		parameters["testing"] = testing
	}

	err := client.CallNoResponseBody(ctx, "nameserver.delete", parameters)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not delete nameserver record",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

// convert soa rname to email
func transformRname(rname string) string {
	// split rname into labels by .
	labels := strings.Split(rname, ".")

	if len(labels) < 2 {
		return rname
	}

	// first part is username the rest is the domain
	username := labels[0]
	domain := strings.Join(labels[1:], ".")

	// unescape "\." in username
	username = strings.ReplaceAll(username, `\.`, ".")

	return username + "@" + domain
}

func resourceNameserverUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createOrUpdateNameserver(ctx, d, m, false)
}
