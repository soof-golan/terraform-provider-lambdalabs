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
	_ datasource.DataSource              = &sshkeyDataSource{}
	_ datasource.DataSourceWithConfigure = &sshkeyDataSource{}
)

// NewSSHKeyDataSource is a helper function to simplify the provider implementation.
func NewSSHKeyDataSource() datasource.DataSource {
	return &sshkeyDataSource{}
}

// sshkeyDataSource is the data source implementation.
type sshkeyDataSource struct {
	client *lambdalabs.ClientWithResponses
}

func (d *sshkeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (d *sshkeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "SSHKey name",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "SSHKey ID",
			},
			"public_key": schema.StringAttribute{
				Computed:    true,
				Description: "SSHKey public key",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *sshkeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *sshkeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sshkeyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state.Name.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unable to Read Lambda Labs SSHKeys",
			"SSHKey name is required",
		)
		return
	}

	r, err := d.client.ListSSHKeysWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lambda Labs SSHKeys",
			err.Error(),
		)
		return
	}
	switch r.StatusCode() {
	case 403, 401:
		resp.Diagnostics.AddError(
			"Unable to Read Lambda Labs SSHKeys",
			"Invalid API Key",
		)
	}
	if r.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Read Lambda Labs SSHKeys",
			"Response body is empty",
		)
		return
	}

	keyIndex := -1
	for index, sshkey := range r.JSON200.Data {
		name := state.Name.ValueString()
		if name == sshkey.Name {
			keyIndex = index
			state.ID = types.StringValue(sshkey.Id)
			state.PublicKey = types.StringValue(sshkey.PublicKey)
			break
		}
	}
	if keyIndex == -1 {
		resp.Diagnostics.AddError(
			"Unable to Read Lambda Labs SSHKeys",
			"SSHKey not found",
		)
		return
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
