package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"
	"terraform-provider-lambdalabs/pgk/lambdalabs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &lambdalabsProvider{}
)

// lambdalabsProviderModel maps provider schema data to a Go type.
type lambdalabsProviderModel struct {
	Host   types.String `tfsdk:"host"`
	ApiKey types.String `tfsdk:"api_key"`
}

// lambdalabsProvider is the provider implementation.
type lambdalabsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *lambdalabsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "lambdalabs"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
// Schema defines the provider-level schema for configuration data.
func (p *lambdalabsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "Lambda Labs API host",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Lambda Labs API key",
			},
		},
	}
}

// Configure prepares a lambdalabs API client for data sources and resources.
func (p *lambdalabsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	tflog.Info(ctx, "Configuring Lambda Labs client")
	var config lambdalabsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Lambda Labs API Host",
			"The provider cannot create the Lambda Labs API client as there is an unknown configuration value for the Lambda Labs API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LAMBDALABS_HOST environment variable.",
		)
	}

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Lambda Labs API Key",
			"The provider cannot create the Lambda Labs API client as there is an unknown configuration value for the Lambda Labs API API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LAMBDALABS_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("LAMBDALABS_HOST")
	apiKey := os.Getenv("LAMBDALABS_API_KEY")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Lambda Labs API Key",
			"The provider cannot create the Lambda Labs API client as there is a missing or empty value for the Lambda Labs API key. "+
				"Set the key value in the configuration or use the LAMBDALABS_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "lambdalabs_host", host)
	ctx = tflog.SetField(ctx, "lambdalabs_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "lambdalabs_api_key")

	tflog.Debug(ctx, "Creating Lambda Labs client")

	// Create a new LambdaLabs client using the configuration values
	lambdaclient, err := lambdalabs.NewAuthenticatedClient(host, apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Lambda Labs API Client",
			"An unexpected error occurred when creating the Lambda Labs API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Lambda Labs Client Error: "+err.Error(),
		)
		return
	}

	// Make the Lambda Labs client available during DataSource and Resource
	// type Configure methods.

	resp.DataSourceData = lambdaclient
	resp.ResourceData = lambdaclient
	tflog.Info(ctx, "Configured Lambda Labs client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *lambdalabsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewInstanceDataSource,
		NewInstancesDataSource,
		NewFilesystemDataSource,
		NewSSHKeysDataSource,
		NewSSHKeyDataSource,
	}
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &lambdalabsProvider{
			version: version,
		}
	}
}

// Resources defines the resources implemented in the provider.
func (p *lambdalabsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInstanceResource,
		NewSSHKeyResource,
	}
}
