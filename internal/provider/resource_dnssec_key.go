// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/inwx/terraform-provider-inwx/internal/api"
	"strconv"
	"strings"
)

type DNSSECKeyResource struct {
	client *api.Client
}

type DNSSECKeyResourceModel struct {
	Id         types.String `tfschema:"id"`
	Domain     types.String `tfschema:"domain"`
	PublicKey  types.String `tfschema:"public_key"`
	Algorithm  types.Int64  `tfschema:"algorithm"`
	Digest     types.String `tfschema:"digest"`
	DigestType types.Int64  `tfschema:"digest_type"`
	Flag       types.Int64  `tfschema:"flag"`
	KeyTag     types.Int64  `tfschema:"key_tag"`
	Status     types.String `tfschema:"status"`
}

func NewDNSSECKeyResource() resource.Resource {
	return &DNSSECKeyResource{}
}

func (r *DNSSECKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnssec_key"
}

func (r *DNSSECKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSSECKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Provides a INWX DNSSEC key resource. This will send your dnssec keys to the domain registry. If you use INWX nameservers, use inwx_automated_dnssec instead, and INWX will create and manage the keys.

## CDS / CDNSKEY

INWX supports CDS for .ch, .li, .se, .nu. If you use this record we will import your keys automatically after a few days.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service generated identifier for dnssec key",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "Name of the domain",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Public key of the domain",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"algorithm": schema.Int64Attribute{
				MarkdownDescription: "Algorithm used for the public key",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"digest": schema.StringAttribute{
				MarkdownDescription: "Computed digest for the public key",
				Computed:            true,
			},
			"digest_type": schema.Int64Attribute{
				MarkdownDescription: "Computed digest type",
				Computed:            true,
			},
			"flag": schema.Int64Attribute{
				MarkdownDescription: "Key flag (256=ZSK, 257=KSK)",
				Computed:            true,
			},
			"key_tag": schema.Int64Attribute{
				MarkdownDescription: "Key tag",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				Description: "DNSSEC status",
				Computed:    true,
			},
		},
	}
}

func (r *DNSSECKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: domain/digest. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("digest"), parts[1])...)
}

func (r *DNSSECKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var data DNSSECKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	parameters := map[string]interface{}{
		"domainName": data.Domain.ValueString(),
		"digest":     data.Digest.ValueString(),
		"active":     1,
	}

	call, err := r.client.Call(ctx, "dnssec.listkeys", parameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not get DNSSEC keys",
			err.Error(),
		)
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		resp.Diagnostics.AddError(
			"Could not get DNSSEC keys",
			fmt.Sprintf("API response not status code 1000. Got response: %s", call.ApiError()),
		)
		return
	}

	resData := call["resData"].([]interface{})
	key := resData[0].(map[string]interface{})

	data.Id = types.StringValue(key["id"].(string))
	data.Domain = types.StringValue(key["ownerName"].(string))
	data.PublicKey = types.StringValue(key["publicKey"].(string))
	data.Digest = types.StringValue(key["digest"].(string))
	data.Status = types.StringValue(key["status"].(string))

	if i, err := strconv.ParseInt(key["algorithmId"].(string), 10, 64); err == nil {
		data.Algorithm = types.Int64Value(i)
	} else {
		resp.Diagnostics.AddError(
			"algorithm: failed to parse int from string",
			err.Error(),
		)
	}

	if i, err := strconv.ParseInt(key["digestTypeId"].(string), 10, 64); err == nil {
		data.DigestType = types.Int64Value(i)
	} else {
		resp.Diagnostics.AddError(
			"digest_type: failed to parse int from string",
			err.Error(),
		)
	}

	if i, err := strconv.ParseInt(key["flagId"].(string), 10, 64); err == nil {
		data.Flag = types.Int64Value(i)
	} else {
		resp.Diagnostics.AddError(
			"flag: failed to parse int from string",
			err.Error(),
		)
	}

	if i, err := strconv.ParseInt(key["keyTag"].(string), 10, 64); err == nil {
		data.KeyTag = types.Int64Value(i)
	} else {
		resp.Diagnostics.AddError(
			"key_tag: failed to parse int from string",
			err.Error(),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSECKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var data DNSSECKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"domainName": data.Domain.ValueString(),
		"dnskey": fmt.Sprintf(
			"%s. IN DNSKEY 257 3 %d %s",
			data.Domain.ValueString(),
			data.Algorithm.ValueInt64(),
			data.PublicKey.ValueString(),
		),
		"calculateDigest": true,
	}

	call, err := r.client.Call(ctx, "dnssec.adddnskey", parameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not add DNSKEY",
			err.Error(),
		)
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError(
			"Could not add DNSKEY",
			fmt.Sprintf("API response not status code 1000 or 1001. Got response: %s", call.ApiError()),
		)
		return
	}

	resData := call["resData"].(map[string]interface{})

	parts := strings.Split(resData["ds"].(string), " ")
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Could not parse returned DS",
			fmt.Sprintf("API response not in expected format. Got response: %s", resData["ds"]),
		)
		return
	}

	data.Digest = types.StringValue(parts[3])

	parameters = map[string]interface{}{
		"domainName": data.Domain.ValueString(),
		"digest":     data.Digest.ValueString(),
		"active":     1,
	}

	call, err = r.client.Call(ctx, "dnssec.listkeys", parameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not get DNSSEC keys",
			err.Error(),
		)
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL {
		resp.Diagnostics.AddError(
			"Could not get DNSSEC keys",
			fmt.Sprintf("API response not status code 1000. Got response: %s", call.ApiError()),
		)
		return
	}

	res := call["resData"].([]interface{})
	key := res[0].(map[string]interface{})

	data.Id = types.StringValue(key["id"].(string))
	data.Domain = types.StringValue(key["ownerName"].(string))
	data.PublicKey = types.StringValue(key["publicKey"].(string))
	data.Digest = types.StringValue(key["digest"].(string))
	data.Status = types.StringValue(key["status"].(string))

	if i, err := strconv.ParseInt(key["algorithmId"].(string), 10, 64); err == nil {
		data.Algorithm = types.Int64Value(i)
	} else {
		resp.Diagnostics.AddError(
			"algorithm: failed to parse int from string",
			err.Error(),
		)
	}

	if i, err := strconv.ParseInt(key["digestTypeId"].(string), 10, 64); err == nil {
		data.DigestType = types.Int64Value(i)
	} else {
		resp.Diagnostics.AddError(
			"digest_type: failed to parse int from string",
			err.Error(),
		)
	}

	if i, err := strconv.ParseInt(key["flagId"].(string), 10, 64); err == nil {
		data.Flag = types.Int64Value(i)
	} else {
		resp.Diagnostics.AddError(
			"flag: failed to parse int from string",
			err.Error(),
		)
	}

	if i, err := strconv.ParseInt(key["keyTag"].(string), 10, 64); err == nil {
		data.KeyTag = types.Int64Value(i)
	} else {
		resp.Diagnostics.AddError(
			"key_tag: failed to parse int from string",
			err.Error(),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSSECKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// NO-OP: Can not update this resource
	resp.Diagnostics.AddWarning(
		"No Update Support",
		"This resource does not support updates. To make changes, please delete and recreate the resource.",
	)

	// Read the current state into the response's State.
	var state NameserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the state back, ensuring nothing has changed.
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *DNSSECKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	var data DNSSECKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	parameters := map[string]interface{}{
		"key": data.Id,
	}

	call, err := r.client.Call(ctx, "dnssec.deletednskey", parameters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not delete DNSKEY",
			err.Error(),
		)
		return
	}
	if call.Code() != api.COMMAND_SUCCESSFUL && call.Code() != api.COMMAND_SUCCESSFUL_PENDING {
		resp.Diagnostics.AddError(
			"Could not delete DNSKEY",
			fmt.Sprintf("API response not status code 1000 pr 1001. Got response: %s", call.ApiError()),
		)
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove the resource from state if there are no other errors.
}
