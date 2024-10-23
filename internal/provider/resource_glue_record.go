package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/inwx/terraform-provider-inwx/internal/api"
	"strconv"
	"strings"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &GlueRecordResource{}
var _ resource.ResourceWithImportState = &GlueRecordResource{}

type GlueRecordResource struct {
	client *api.Client
}

func NewGlueRecordResource() resource.Resource {
	return &GlueRecordResource{}
}

func (r *GlueRecordResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "inwx_glue_record"
}

func (r *GlueRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GlueRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: "The name of the host.",
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ro_id": schema.Int64Attribute{
				Description: "Repository Object Identifier (RO ID) of the hostname.",
				Required:    true,
			},
			"ip": schema.ListAttribute{
				Description: "List of IP addresses associated with the hostname.",
				ElementType: types.StringType,
				Required:    true,
			},
			"testing": schema.BoolAttribute{
				Description: "Execute the command in testing mode.",
				Optional:    true,
			},
		},
	}
}

func (r *GlueRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var plan GlueRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"hostname": plan.Hostname.ValueString(),
		"ip":       plan.Ip.Elements(),
	}

	if plan.Testing.ValueBool() {
		parameters["testing"] = plan.Testing.ValueBool()
	}

	call, err := r.client.Call(ctx, "host.create", parameters)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not create glue record: %s", err))
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unexpected API response: %s", call.ApiError()))
		return
	}

	resData := call["resData"].(map[string]any)
	roID := strconv.Itoa(int(resData["roId"].(float64)))

	plan.RoID = types.Int64Value(int64(resData["roId"].(float64)))
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.Hostname.ValueString(), roID))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GlueRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var state GlueRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"hostname": state.Hostname.ValueString(),
	}

	call, err := r.client.Call(ctx, "host.info", parameters)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not read glue record: %s", err))
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unexpected API response: %s", call.ApiError()))
		return
	}

	records := call["resData"].(map[string]any)["record"].([]any)
	for _, rec := range records {
		record := rec.(map[string]any)
		recordID := fmt.Sprintf("%s:%s", state.Hostname.ValueString(), strconv.Itoa(int(record["roId"].(float64))))
		if recordID == state.ID.ValueString() {
			state.RoID = types.Int64Value(int64(record["roId"].(float64)))
			state.Hostname = types.StringValue(record["hostname"].(string))
			state.Ip, _ = types.ListValueFrom(ctx, types.StringType, record["ip"].([]any))
		}
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *GlueRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var plan GlueRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, roID, err := parseGlueRecordID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing ID", fmt.Sprintf("Could not parse ID: %s", err))
		return
	}

	parameters := map[string]interface{}{
		"roId": roID,
		"ip":   plan.Ip.Elements(),
	}
	if plan.Hostname.IsUnknown() {
		parameters["hostname"] = plan.Hostname.ValueString()
	}
	if plan.Testing.IsUnknown() {
		parameters["testing"] = plan.Testing.ValueBool()
	}

	err = r.client.CallNoResponseBody(ctx, "host.update", parameters)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not update glue record: %s", err))
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GlueRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var state GlueRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, roID, err := parseGlueRecordID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing ID", fmt.Sprintf("Could not parse ID: %s", err))
		return
	}

	parameters := map[string]interface{}{
		"roId": roID,
	}

	err = r.client.CallNoResponseBody(ctx, "host.delete", parameters)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not delete glue record: %s", err))
		return
	}
}

func (r *GlueRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	domain, id, err := parseGlueRecordID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Parsing ID", fmt.Sprintf("Could not parse ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hostname"), domain)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ro_id"), id)...)
}

// GlueRecordResourceModel defines the data structure for the resource's state.
type GlueRecordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Hostname types.String `tfsdk:"hostname"`
	RoID     types.Int64  `tfsdk:"ro_id"`
	Ip       types.List   `tfsdk:"ip"`
	Testing  types.Bool   `tfsdk:"testing"`
}

// parseGlueRecordID parses the ID format expected in the resource.
func parseGlueRecordID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected 'attribute1:attribute2'", id)
	}
	return parts[0], parts[1], nil
}
