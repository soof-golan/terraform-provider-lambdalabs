package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-lambdalabs/pgk/lambdalabs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &instancesDataSource{}
	_ datasource.DataSourceWithConfigure = &instancesDataSource{}
)

// NewInstancesDataSource is a helper function to simplify the provider implementation.
func NewInstancesDataSource() datasource.DataSource {
	return &instancesDataSource{}
}

// instancesDataSource is the data source implementation.
type instancesDataSource struct {
	client *lambdalabs.ClientWithResponses
}

// Metadata returns the data source type name.
func (d *instancesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instances"
}

// Schema defines the data source schema.
func (d *instancesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instances": schema.ListNestedAttribute{
				MarkdownDescription: "#Instances\n\nList of instances",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"filesystem_names": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "#FilesystemNames\n\nList of filesystem names attached to this instance",
						},
						"hostname": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "# Hostname\n\nassigned to this instance, which resolves to the instance's IP.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "#Id\n\nUnique identifier of the instance",
						},
						"instance_type": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "#InstanceType\n\nHardware configuration and pricing of an instance type",
							Attributes: map[string]schema.Attribute{
								"description": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "#Description\n\nLong name of the instance type",
								},
								"name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "#Name\n\nName of an instance type",
								},
								"price_cents_per_hour": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "#PriceCentsPerHour\n\nPrice of the instance type, in US cents per hour",
								},
								"specs": schema.SingleNestedAttribute{
									Computed:            true,
									MarkdownDescription: "#Specs\n\nHardware configuration of an instance type",
									Attributes: map[string]schema.Attribute{
										"memory_gib": schema.Int64Attribute{
											Computed:            true,
											MarkdownDescription: "#MemoryGib\n\nAmount of RAM, in gibibytes (GiB)",
										},
										"storage_gib": schema.Int64Attribute{
											Computed:            true,
											MarkdownDescription: "#StorageGib\n\nAmount of storage, in gibibytes (GiB).",
										},
										"vcpus": schema.Int64Attribute{
											Computed:            true,
											MarkdownDescription: "#Vcpus\n\nNumber of virtual CPUs",
										},
									},
								},
							},
						},
						"ip": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "# IP\n\nIPv4 address of the instance",
						},
						"jupyter_token": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "# JupyterToken\n\nSecret token used to log into the jupyter lab server hosted on the instance.",
						},
						"jupyter_url": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "# JupyterUrl\n\nURL that opens a jupyter lab notebook on the instance.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Optional:            true,
							MarkdownDescription: "#Name\n\nName of the instance",
						},
						"region": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "#Region\n\nRegion of the instance",
							Attributes: map[string]schema.Attribute{
								"description": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "#Description\n\nLong name of the region",
								},
								"name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "#Name\n\nName of the region",
								},
							},
						},
						"ssh_key_names": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "#SSHKeyNames\n\nNames of the SSH keys allowed to access the instance",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "#Status\n\nThe current status of the instance",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *instancesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*lambdalabs.ClientWithResponses)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *lambdalabs.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *instancesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state InstancesDataSourceModel

	if resp.Diagnostics.HasError() {
		return
	}

	r, err := d.client.ListInstancesWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lambda Labs Instances. Lambda Labs Client Error",
			err.Error(),
		)
		return
	}
	if r.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lambda Labs Instances",
			string(r.Body),
		)
		return
	}

	for _, instance := range r.JSON200.Data {
		instanceState := InstanceDataSourceModel{
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

		state.Instances = append(state.Instances, instanceState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
