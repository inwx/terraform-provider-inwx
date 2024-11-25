package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/inwx/terraform-provider-inwx/internal/api"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NameserverRecordResource represents the nameserver record resource
type NameserverRecordResource struct {
	client *api.Client
}

// NewNameserverRecordResource creates a new instance of the resource
func NewNameserverRecordResource() resource.Resource {
	return &NameserverRecordResource{}
}

// Metadata provides resource type name
func (r *NameserverRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameserver_record"
}

// Schema defines the resource schema
func (r *NameserverRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	validRecordTypes := []string{
		"A", "AAAA", "AFSDB", "ALIAS", "CAA", "CERT", "CNAME", "HINFO", "KEY", "LOC", "MX", "NAPTR", "NS", "OPENPGPKEY",
		"PTR", "RP", "SMIMEA", "SOA", "SRV", "SSHFP", "TLSA", "TXT", "URI", "URL",
	}

	validUrlRedirectTypes := []string{
		"HEADER301", "HEADER302", "FRAME",
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "Domain name",
				Required:    true,
			},
			"ro_id": schema.Int64Attribute{
				Description: "DNS domain ID",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the nameserver record. One of: " + strings.Join(validRecordTypes, ", "),
				Validators: []validator.String{
					stringvalidator.OneOf(validRecordTypes...),
				},
				Required: true,
			},
			"content": schema.StringAttribute{
				Description: "Content of the nameserver record",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the nameserver record",
				Optional:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "TTL (time to live) of the nameserver record",
				Optional:    true,
				Default:     int64default.StaticInt64(3600),
			},
			"prio": schema.Int64Attribute{
				Description: "Priority of the nameserver record",
				Optional:    true,
				Default:     int64default.StaticInt64(0),
			},
			"url_redirect_type": schema.StringAttribute{
				Description: "Type of the URL redirection. One of: " + strings.Join(validUrlRedirectTypes, ", "),
				Validators: []validator.String{
					stringvalidator.OneOf(validUrlRedirectTypes...),
				},
				Optional: true,
			},
			"url_redirect_title": schema.StringAttribute{
				Description: "Title of the frame redirection",
				Optional:    true,
			},
			"url_redirect_description": schema.StringAttribute{
				Description: "Description of the frame redirection",
				Optional:    true,
			},
			"url_redirect_fav_icon": schema.StringAttribute{
				Description: "FavIcon of the frame redirection",
				Optional:    true,
			},
			"url_redirect_keywords": schema.StringAttribute{
				Description: "Keywords of the frame redirection",
				Optional:    true,
			},
			"url_append": schema.BoolAttribute{
				Description: "Append the path for redirection",
				Optional:    true,
			},
			"testing": schema.BoolAttribute{
				Description: "Execute command in testing mode",
				Optional:    true,
			},
		},
	}
}

type NameserverRecordModel struct {
	ID                     types.String `tfsdk:"id"`
	Domain                 types.String `tfsdk:"domain"`
	RoID                   types.Int64  `tfsdk:"ro_id"`
	Type                   types.String `tfsdk:"type"`
	Content                types.String `tfsdk:"content"`
	Name                   types.String `tfsdk:"name"`
	TTL                    types.Int64  `tfsdk:"ttl"`
	Priority               types.Int64  `tfsdk:"prio"`
	URLRedirectType        types.String `tfsdk:"url_redirect_type"`
	URLRedirectTitle       types.String `tfsdk:"url_redirect_title"`
	URLRedirectDescription types.String `tfsdk:"url_redirect_description"`
	URLRedirectFavIcon     types.String `tfsdk:"url_redirect_fav_icon"`
	URLRedirectKeywords    types.String `tfsdk:"url_redirect_keywords"`
	URLAppend              types.Bool   `tfsdk:"url_append"`
	Testing                types.Bool   `tfsdk:"testing"`
}

// Create handles the resource creation
func (r *NameserverRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NameserverRecordModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"domain":  data.Domain.ValueString(),
		"type":    data.Type.ValueString(),
		"content": data.Content.ValueString(),
	}

	// Optional attributes
	if !data.RoID.IsNull() {
		parameters["roId"] = data.RoID.ValueInt64()
	}
	if !data.Name.IsNull() {
		parameters["name"] = data.Name.ValueString()
	}
	if !data.TTL.IsNull() {
		parameters["ttl"] = data.TTL.ValueInt64()
	}
	if !data.Priority.IsNull() {
		parameters["prio"] = data.Priority.ValueInt64()
	}

	call, err := r.client.Call(ctx, "nameserver.createRecord", parameters)
	if err != nil {
		resp.Diagnostics.AddError("Create failed", err.Error())
		return
	}

	if call.Code() != api.COMMAND_SUCCESSFUL {
		resp.Diagnostics.AddError("API Error", call.ApiError())
		return
	}

	resData := call["resData"].(map[string]interface{})
	data.ID = types.StringValue(fmt.Sprintf("%s:%d", data.Domain.ValueString(), int(resData["id"].(float64))))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Read handles resource reading
func (r *NameserverRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NameserverRecordModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters := map[string]interface{}{
		"domain": data.Domain.ValueString(),
	}

	call, err := r.client.Call(ctx, "nameserver.info", parameters)
	if err != nil {
		resp.Diagnostics.AddError("Read failed", err.Error())
		return
	}

	records := call["resData"].(map[string]interface{})["record"].([]interface{})
	for _, record := range records {
		recordData := record.(map[string]interface{})
		if fmt.Sprintf("%s:%d", data.Domain.ValueString(), int(recordData["id"].(float64))) == data.ID.ValueString() {
			data.Type = types.StringValue(recordData["type"].(string))
			data.Content = types.StringValue(recordData["content"].(string))
			diags = resp.State.Set(ctx, &data)
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

// Update handles the resource update
func (r *NameserverRecordResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported Operation", "Updating nameserver records is not supported.")
}

// Delete handles the resource deletion
func (r *NameserverRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NameserverRecordModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, id, err := resourceNameserverRecordParseId(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete failed", "Invalid ID format")
		return
	}

	parameters := map[string]interface{}{
		"id": id,
	}

	err = r.client.CallNoResponseBody(ctx, "nameserver.deleteRecord", parameters)
	if err != nil {
		resp.Diagnostics.AddError("Delete failed", err.Error())
	}
}

// resourceNameserverRecordParseId parses the ID of the nameserver record into domain and record ID.
func resourceNameserverRecordParseId(id string) (domain string, recordID string, err error) {
	// Split the ID into its parts
	parts := strings.Split(id, ":")

	// Check if the format is correct (should contain exactly two parts)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected 'domain:id'", id)
	}

	// Assign parts to domain and recordID
	domain = parts[0]
	recordID = parts[1]

	return domain, recordID, nil
}
