package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/inwx/terraform-provider-inwx/internal/api"
	"net/url"
	"os"

	"github.com/go-logr/logr"
)

type inwxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type inwxProviderModel struct {
	ApiUrl   types.String `tfsdk:"api_url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Tan      types.String `tfsdk:"tan"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &inwxProvider{
			version: version,
		}
	}
}

func (p *inwxProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "inwx"
	resp.Version = p.version
}

func (p *inwxProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The INWX Provider can be used to register and manage domains and their domain contacts. Additionally it offers full support for nameserver and DNSSEC management.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Description: "URL of the RPC API endpoint. Use `https://api.domrobot.com/jsonrpc/` " +
					"for production and `https://api.ote.domrobot.com/jsonrpc/` for tests. " +
					"Can be passed as `INWX_API_URL` env var.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				Description: "Login username of the api. Can be passed as `INWX_USERNAME` env var.",
				Required:    true,
				Sensitive:   true,
			},
			"password": schema.StringAttribute{
				Description: "Login password of the api. Can be passed as `INWX_PASSWORD` env var.",
				Required:    true,
				Sensitive:   true,
			},
			"tan": schema.StringAttribute{
				Description: "Mobile-TAN to unlock account. Can be passed as `INWX_TAN` env var.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func GetEnvDefault(key, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

func (p *inwxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config inwxProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the attributes, it must be a known value.

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown INWX API Username",
			"The provider cannot create the INWX API client as there is an unknown configuration value for the INWX API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the INWX_USERNAME environment variable.",
		)
	}
	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown INWX API Password",
			"The provider cannot create the INWX API client as there is an unknown configuration value for the INWX API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the INWX_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override with Terraform configuration value if set.

	apiUrl := GetEnvDefault("INWX_API_URL", "https://api.domrobot.com/jsonrpc/")
	username := os.Getenv("INWX_USERNAME")
	password := os.Getenv("INWX_PASSWORD")
	tan := GetEnvDefault("INWX_TAN", "")

	if !config.ApiUrl.IsNull() {
		apiUrl = config.ApiUrl.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.Tan.IsNull() {
		tan = config.Tan.ValueString()
	}

	// If any of the expected configurations are missing, return errors with provider-specific guidance.

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing INWX API Username",
			"The provider cannot create the INWX API client as there is a missing or empty value for the INWX API username. "+
				"Set the username value in the configuration or use the INWX_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing INWX API Password",
			"The provider cannot create the INWX API client as there is a missing or empty value for the INWX API password. "+
				"Set the password value in the configuration or use the INWX_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	papiUrl, err := url.Parse(apiUrl)

	if err != nil {
		resp.Diagnostics.AddError(
			"Could not configure context",
			fmt.Sprintf("Could not parse api_url: %w", err),
		)
		return
	}

	logger := logr.Discard()

	client, err := api.NewClient(username, password, papiUrl, &logger, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create INWX API Client",
			"An unexpected error occurred when creating the INWX API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"INWX Client Error: "+err.Error(),
		)
		return
	}

	loginParams := map[string]interface{}{
		"user": username,
		"pass": password,
	}
	call, err := client.Call(ctx, "account.login", loginParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not configure context",
			fmt.Sprintf("Could not authenticate at api via account.login: %w", err),
		)
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		resp.Diagnostics.AddError(
			"Could not configure context",
			fmt.Sprintf("Could not authenticate at api via account.login. "+
				"Got response: %s", call.ApiError()),
		)
		return
	}

	call, err = client.Call(ctx, "account.info", map[string]interface{}{})

	if call != nil && call.Code() == 2200 && tan != "" {
		call, err := client.Call(ctx, "account.unlock", map[string]interface{}{
			"tan": tan,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not unlock account",
				fmt.Sprintf("Could not authenticate at api via account.unlock: %w", err),
			)
			return
		}
		if call.Code() != api.COMMAND_SUCCESSFUL {
			resp.Diagnostics.AddError(
				"Could not unlock account",
				fmt.Sprintf("Could not authenticate at api via account.unlock. "+
					"Got response: %s", call.ApiError()),
			)
			return
		}
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *inwxProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainContactDataSource,
	}
}

func (p *inwxProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAutomatedDNSSECResource,
		NewDNSSECKeyResource,
		NewGlueRecordResource,
		NewNameserverResource,
		NewDomainResource,
		NewDomainContactResource,
		NewNameserverRecordResource,
	}
}
