package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/inwx/terraform-provider-inwx/internal/api"
)

type automatedDNSSECResource struct {
	client *api.Client
}

type automatedDNSSECResourceModel struct {
	Domain types.String `tfsdk:"domain"`
	Id     types.String `tfsdk:"id"`
}

func NewAutomatedDNSSECResource() resource.Resource {
	return &automatedDNSSECResource{}
}

func (r *automatedDNSSECResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automated_dnssec"
}

func (r *automatedDNSSECResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *automatedDNSSECResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "Name of the domain",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service generated identifier for dnssec",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		MarkdownDescription: "Automated DNSSEC management for a domain.",
	}
}

func (r *automatedDNSSECResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var data automatedDNSSECResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	parameters := map[string]interface{}{
		"domains": []string{data.Domain.ValueString()},
	}

	call, err := r.client.Call(ctx, "dnssec.info", parameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read DNSSEC info",
			err.Error(),
		)
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError(
			"Could not read DNSSEC info",
			fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		)
		return
	}

	records := call["resData"].(map[string]any)["data"].([]any)

	for _, record := range records {
		recordt := record.(map[string]any)

		if recordt["domain"].(string) == data.Domain.ValueString() && recordt["dnssecStatus"].(string) == "AUTO" {
			data.Id = types.StringValue(recordt["domain"].(string))
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *automatedDNSSECResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var data automatedDNSSECResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"domainName": data.Domain.ValueString(),
	}

	call, err := r.client.Call(ctx, "dnssec.enablednssec", parameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not enable automated DNSSEC",
			err.Error(),
		)
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError(
			"Could not enable automated DNSSEC",
			fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		)
		return
	}

	data.Id = types.StringValue(data.Domain.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *automatedDNSSECResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// NO-OP: Can not update this resource
}

func (r *automatedDNSSECResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var data automatedDNSSECResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	parameters := map[string]interface{}{
		"domainName": data.Domain.ValueString(),
	}

	call, err := r.client.Call(ctx, "dnssec.disablednssec", parameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not disable automated DNSSEC",
			err.Error(),
		)
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError(
			"Could not disable automated DNSSEC",
			fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		)
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove the resource from state if there are no other errors.
}
