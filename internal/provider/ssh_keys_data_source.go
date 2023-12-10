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
	_ datasource.DataSource              = &sshkeysDataSource{}
	_ datasource.DataSourceWithConfigure = &sshkeysDataSource{}
)

// NewSSHKeysDataSource is a helper function to simplify the provider implementation.
func NewSSHKeysDataSource() datasource.DataSource {
	return &sshkeysDataSource{}
}

// sshkeysDataSource is the data source implementation.
type sshkeysDataSource struct {
	client *lambdalabs.ClientWithResponses
}

// sshkeysDataSourceModel maps the data source schema data.
type sshkeysDataSourceModel struct {
	SSHKeys []sshkeyModel `tfsdk:"sshkeys"`
}

// sshkeyModel maps sshkey schema data.
type sshkeyModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	PublicKey types.String `tfsdk:"public_key"`
}

func (d *sshkeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keys"
}

func (d *sshkeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"sshkeys": schema.ListNestedAttribute{
				Description: "List of All SSH Keys",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "SSHKey ID",
							MarkdownDescription: `SSH Key ID (unique within account)`,
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "SSH Key name",
							MarkdownDescription: `SSH Key Name (unique within account)`,
						},
						"public_key": schema.StringAttribute{
							Computed:            true,
							Description:         "SSHKey public key",
							MarkdownDescription: `SSH Key Public Key. used to access VMs.`,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *sshkeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

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
func (d *sshkeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sshkeysDataSourceModel

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

	sshkeys := r.JSON200.Data

	// Map response body to model
	for _, sshkey := range sshkeys {

		sshkeyState := sshkeyModel{
			ID:        types.StringValue(sshkey.Id),
			Name:      types.StringValue(sshkey.Name),
			PublicKey: types.StringValue(sshkey.PublicKey),
		}
		state.SSHKeys = append(state.SSHKeys, sshkeyState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
