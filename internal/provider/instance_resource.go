// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"terraform-provider-lambdalabs/pgk/lambdalabs"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &InstanceResource{}
var _ resource.ResourceWithConfigure = &InstanceResource{}
var _ resource.ResourceWithImportState = &InstanceResource{}

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

// InstanceResource defines the resource implementation.
type InstanceResource struct {
	client *lambdalabs.ClientWithResponses
}

func (r *InstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Lambda Labs VM Instance resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the instance. valid when `quantity` is 1 (the default).",
			},
			//"ids": schema.SetAttribute{
			//	Computed:            true,
			//	MarkdownDescription: `List of unique identifiers of the instances`,
			//	ElementType:         types.StringType,
			//},
			"name": schema.StringAttribute{
				MarkdownDescription: "User-provided name of the instance",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssh_key_names": schema.ListAttribute{
				MarkdownDescription: "List of SSH Key names to be added to the instance. Currently, exactly one SSH key must be specified.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					ListMaxLength{max: 1},
				},
				PlanModifiers: []planmodifier.List{
					// Must redeploy because SSH keys cannot be detached while instance is running
					listplanmodifier.RequiresReplace(),
				},
			},
			"filesystem_names": schema.ListAttribute{
				MarkdownDescription: "List of filesystem names to be added to the instance. Currently, only one (if any) file system may be specified.",
				ElementType:         types.StringType,
				Computed:            true,
				Default: ListDefaultEmpty{
					ElementType: types.StringType,
				},
				Validators: []validator.List{
					ListMaxLength{max: 1},
				},
				PlanModifiers: []planmodifier.List{
					// Must redeploy because filesystems cannot be detached while instance is running
					listplanmodifier.RequiresReplace(),
				},
			},
			"instance_type": schema.StringAttribute{
				MarkdownDescription: "Name of an instance type",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					// Must redeploy if instance type changes
					stringplanmodifier.RequiresReplace(),
				},
			},
			//"quantity": schema.Int64Attribute{
			//	MarkdownDescription: "Number of instances to provision (default: 1)",
			//	Optional:            true,
			//	Computed:            true,
			//	Default:             Int64Default{defaultValue: 1},
			//	PlanModifiers: []planmodifier.Int64{
			//		int64planmodifier.RequiresReplace(),
			//	},
			//},
			"region": schema.StringAttribute{
				MarkdownDescription: "Name of the region where the instance is located",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					// Must redeploy if region changes
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceResourceModel

	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(ctx, 20*time.Minute, func() *retry.RetryError {
		instanceTypesResponse, err := r.client.InstanceTypesWithResponse(ctx)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("unable to read available instance types, got error: %s", err))
		}
		if instanceTypesResponse.JSON200 == nil {
			return retry.NonRetryableError(fmt.Errorf("unable to read available instance types, got error: %s", instanceTypesResponse.Body))
		}
		instanceTypes := instanceTypesResponse.JSON200.Data
		instanceAvailability, ok := instanceTypes[data.InstanceTypeName.ValueString()]
		if !ok {
			return retry.NonRetryableError(fmt.Errorf("instance type %s not found in available instance types", data.InstanceTypeName.ValueString()))
		}

		regions := make(map[string]lambdalabs.Region)
		for _, region := range instanceAvailability.RegionsWithCapacityAvailable {
			regions[region.Name] = region
		}
		_, ok = regions[data.RegionName.ValueString()]
		if !ok {
			// https://docs.lambdalabs.com/cloud/rate-limiting/
			// Default retry delay is 500ms, so we sleep for 2 more seconds to play nice with the API
			time.Sleep(2 * time.Second)
			return retry.RetryableError(fmt.Errorf("no capacity available in region %s", data.RegionName.ValueString()))
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Instance type unavailable", fmt.Sprintf("Unable to create instance, got error: %s", err))
		return
	}

	var fileSystemNames = make([]string, 0)
	if data.FileSystemNames != nil {
		fileSystemNames = makeStringListFromTf(data.FileSystemNames)
	}

	// TODO: support multiple instances
	var quantity = 1

	body := lambdalabs.LaunchInstanceJSONRequestBody{
		FileSystemNames:  &fileSystemNames,
		InstanceTypeName: data.InstanceTypeName.ValueString(),
		Name:             data.Name.ValueStringPointer(),
		Quantity:         &quantity,
		RegionName:       data.RegionName.ValueString(),
		SshKeyNames:      makeStringListFromTf(data.SshKeyNames),
	}

	response, err := r.client.LaunchInstanceWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("HTTP Client Error", fmt.Sprintf("Unable to create instance, got error: %s", err))
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to create instance. HTTP Status Code: %d", response.StatusCode()),
			fmt.Sprintf("Unable to create instance, got error: %s", response.Body),
		)
		return
	}

	if response.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Failed to create instance (no HTTP 200 response body)",
			fmt.Sprintf("Unable to create instance, got error: %s", response.Body),
		)
		return
	}

	InstanceIDs := response.JSON200.Data.InstanceIds

	if len(InstanceIDs) == 1 {
		tflog.Trace(ctx, "created new instance", map[string]interface{}{"id": InstanceIDs[0]})
		data.ID = types.StringValue(response.JSON200.Data.InstanceIds[0])
	} else {
		tflog.Trace(ctx, "created new instances", map[string]interface{}{"ids": InstanceIDs})
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InstanceResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("reading current instance state %s", state.ID))
	response, err := r.client.ListInstancesWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read instances %s, got error: %s", state.ID, err))
		return
	}
	if response.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Failed to read instances (no HTTP 200 response body)",
			fmt.Sprintf("Unable to read instances %s, got response: %s", state.ID, response.Body),
		)
		return
	}

	var instances = make(map[string]InstanceDataSourceModel)
	for _, instance := range response.JSON200.Data {
		instances[instance.Id] = InstanceDataSourceModel{
			ID:       types.StringValue(instance.Id),
			Hostname: types.StringPointerValue(instance.Hostname),
			Ip:       types.StringPointerValue(instance.Ip),
			Name:     types.StringPointerValue(instance.Name),
			Region: &RegionModel{
				Name:        types.StringValue(instance.Region.Name),
				Description: types.StringValue(instance.Region.Description),
			},
			FileSystemNames: makeTfStringList(instance.FileSystemNames),
			InstanceType: &InstanceTypeModel{
				Description:       types.StringValue(instance.InstanceType.Description),
				Name:              types.StringValue(instance.InstanceType.Name),
				PriceCentsPerHour: types.Int64Value(int64(instance.InstanceType.PriceCentsPerHour)),
				Specs: InstanceSpecsModel{
					MemoryGib:  types.Int64Value(int64(instance.InstanceType.Specs.MemoryGib)),
					StorageGib: types.Int64Value(int64(instance.InstanceType.Specs.StorageGib)),
					Vcpus:      types.Int64Value(int64(instance.InstanceType.Specs.Vcpus)),
				},
			},
			JupyterToken: types.StringPointerValue(instance.JupyterToken),
			JupyterUrl:   types.StringPointerValue(instance.JupyterUrl),
			SshKeyNames:  makeTfStringList(instance.SshKeyNames),
			Status:       types.StringValue(string(instance.Status)),
		}
	}

	instance, ok := instances[state.ID.ValueString()]
	if ok {
		state.Name = instance.Name
		state.RegionName = instance.Region.Name
		state.FileSystemNames = instance.FileSystemNames
		state.InstanceTypeName = instance.InstanceType.Name
		state.SshKeyNames = instance.SshKeyNames
	} else {
		resp.Diagnostics.AddError(
			"Failed to read instance",
			fmt.Sprintf("Unable to read instance %s, got error: %s", state.ID, response.Body),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: support in-place updates when the API supports it
	resp.Diagnostics.AddError("Update not supported", "Lambda Labs Instance resource does not support updates. please report this issue to the provider developers.")
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	instanceId := data.ID.ValueString()
	body := lambdalabs.TerminateInstanceJSONRequestBody{
		InstanceIds: []string{
			instanceId,
		},
	}

	tflog.Debug(ctx, fmt.Sprintf("terminating instances %s", data.ID))
	response, err := r.client.TerminateInstanceWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("HTTP Client Error", fmt.Sprintf("Unable to delete instances %s, got error: %s", data.ID, err))
		return
	}
	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"HTTP Client Error, HTTP Status Code: %d",
			fmt.Sprintf("Unable to delete instances %s, got error: %s", data.ID, response.Body),
		)
		return
	}
	if response.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Failed to delete instances (no HTTP 200 response body)",
			fmt.Sprintf("Unable to delete instances %s, got error: %s", data.ID, response.Body),
		)
		return
	}

	terminated := response.JSON200.Data.TerminatedInstances
	terminatedIDs := make([]string, 0)
	for _, instance := range terminated {
		terminatedIDs = append(terminatedIDs, instance.Id)
	}

	if len(terminatedIDs) == 0 {
		resp.Diagnostics.AddError(
			"Failed to delete instances",
			fmt.Sprintf("Unable to delete instances %s, got error: %s", data.ID, response.Body),
		)
	}

	if len(terminatedIDs) == 1 {
		tflog.Info(ctx, "Terminated instance", map[string]interface{}{"id": terminatedIDs[0]})
	} else {
		tflog.Info(ctx, "Terminated instances", map[string]interface{}{"ids": terminatedIDs})
	}
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("ids"), req, resp)
}
