// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-lambdalabs/pgk/lambdalabs"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SshKeyResource{}
var _ resource.ResourceWithConfigure = &SshKeyResource{}
var _ resource.ResourceWithImportState = &SshKeyResource{}

func NewSSHKeyResource() resource.Resource {
	return &SshKeyResource{}
}

// SshKeyResource defines the resource implementation.
type SshKeyResource struct {
	client *lambdalabs.ClientWithResponses
}

func (r *SshKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *SshKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SSH Key resource Lambda Labs VMs.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: `SSH Key ID (read-only)`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SSH Key Name (must be unique within the account)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "SSH Key Public Key",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *SshKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*lambdalabs.ClientWithResponses)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *lambdalabs.ClientWithResponses, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SshKeyResourceModel

	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body := lambdalabs.AddSSHKeyJSONRequestBody{
		Name:      data.Name.ValueString(),
		PublicKey: data.PublicKey.ValueStringPointer(),
	}
	response, err := r.client.AddSSHKeyWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create ssh_key, got error: %s", err))
		return
	}
	if response.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Failed to create SSH Key",
			fmt.Sprintf("Unable to create ssh_key, got error: %s", response.Body),
		)
		return
	}

	data.Name = types.StringValue(response.JSON200.Data.Name)
	data.PublicKey = types.StringValue(response.JSON200.Data.PublicKey)
	data.Id = types.StringValue(response.JSON200.Data.Id)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SshKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.ListSSHKeysWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ssh_key, got error: %s", err))
		return
	}
	if response.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Failed to read SSH Key",
			fmt.Sprintf("Unable to read ssh_key, got error: %s", response.Body),
		)
		return
	}

	keyIndex := -1
	for index, sshKey := range response.JSON200.Data {
		if data.Name.ValueString() == sshKey.Name {
			keyIndex = index
			data.Name = types.StringValue(sshKey.Name)
			data.PublicKey = types.StringValue(sshKey.PublicKey)
			break
		}
	}
	if keyIndex == -1 {
		tflog.Trace(ctx, "SSH Key not found")
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "SSH Key resource does not support updates. please report this issue to the provider developers.")
}

func (r *SshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SshKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.DeleteSSHKeyWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete ssh_key, got error: %s", err))
		return
	}
	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Failed to delete SSH Key",
			fmt.Sprintf("Unable to delete ssh_key, got error: %s", response.Body),
		)
		return
	}

	tflog.Info(ctx, "Deleted SSH Key", map[string]interface{}{"id": data.Id})
}

func (r *SshKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
