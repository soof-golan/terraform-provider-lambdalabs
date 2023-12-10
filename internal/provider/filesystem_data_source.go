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
	_ datasource.DataSource              = &filesystemsDataSource{}
	_ datasource.DataSourceWithConfigure = &filesystemsDataSource{}
)

// NewFilesystemDataSource is a helper function to simplify the provider implementation.
func NewFilesystemDataSource() datasource.DataSource {
	return &filesystemsDataSource{}
}

// filesystemsDataSource is the data source implementation.
type filesystemsDataSource struct {
	client *lambdalabs.ClientWithResponses
}

// filesystemsDataSourceModel maps the data source schema data.
type filesystemsDataSourceModel struct {
	Filesystems []filesystemModel `tfsdk:"filesystems"`
}

func (d *filesystemsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filesystems"
}

func (d *filesystemsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filesystems": schema.ListNestedAttribute{
				Description: "List of filesystems",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Filesystem ID",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Filesystem name",
						},
						"is_in_use": schema.BoolAttribute{
							Computed:    true,
							Description: "Is the filesystem in use",
						},
						"bytes_used": schema.Int64Attribute{
							Computed:    true,
							Description: "Bytes used",
						},
						"created": schema.StringAttribute{
							Computed:    true,
							Description: "Filesystem creation date",
						},
						"region": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Filesystem region",
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "Filesystem region name",
								},
								"description": schema.StringAttribute{
									Computed:    true,
									Description: "Filesystem region description",
								},
							},
						},
						"mount_point": schema.StringAttribute{
							Computed:    true,
							Description: "Filesystem mount point",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *filesystemsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

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
func (d *filesystemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state filesystemsDataSourceModel

	response, err := d.client.ListFileSystemsWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read file systems. Lambda Labs Client Error",
			err.Error(),
		)
		return
	}
	if response.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lambda Labs Filesystems",
			string(response.Body),
		)
		return
	}

	// Map response body to model
	for _, filesystem := range response.JSON200.Data {
		filesystemState := filesystemModel{
			ID:        types.StringValue(filesystem.Id),
			Name:      types.StringValue(filesystem.Name),
			IsInUse:   types.BoolValue(filesystem.IsInUse),
			BytesUsed: makeOptionalTfInt64(filesystem.BytesUsed),
			Created:   types.StringValue(filesystem.Created),
			Region: RegionModel{
				Name:        types.StringValue(filesystem.Region.Name),
				Description: types.StringValue(filesystem.Region.Description),
			},
			MountPoint: types.StringValue(filesystem.MountPoint),
		}
		state.Filesystems = append(state.Filesystems, filesystemState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
