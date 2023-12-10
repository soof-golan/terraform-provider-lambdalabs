package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RegionModel Region where an instance (or filesystem) is located
type RegionModel struct {
	// Description Long name of the region
	Description types.String `tfsdk:"description"`

	// Name Name of the region
	Name types.String `tfsdk:"name"`
}

// SshKeyResourceModel describes the resource data model.
type SshKeyResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	PublicKey types.String `tfsdk:"public_key"`
}

// filesystemModel maps filesystem schema data.
type filesystemModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	IsInUse    types.Bool   `tfsdk:"is_in_use"`
	BytesUsed  types.Int64  `tfsdk:"bytes_used"`
	Created    types.String `tfsdk:"created"`
	Region     RegionModel  `tfsdk:"region"`
	MountPoint types.String `tfsdk:"mount_point"`
}

// sshkeyDataSourceModel maps the data source schema data.
type sshkeyDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	PublicKey types.String `tfsdk:"public_key"`
}

// InstanceSpecsModel Hardware configuration of an instance type
type InstanceSpecsModel struct {
	// MemoryGib Amount of RAM, in gibibytes (GiB)
	MemoryGib types.Int64 `tfsdk:"memory_gib"`

	// StorageGib Amount of storage, in gibibytes (GiB).
	StorageGib types.Int64 `tfsdk:"storage_gib"`

	// Vcpus Number of virtual CPUs
	Vcpus types.Int64 `tfsdk:"vcpus"`
}

// InstanceTypeModel Hardware configuration and pricing of an instance type
type InstanceTypeModel struct {
	// Description Long name of the instance type
	Description types.String `tfsdk:"description"`

	// Name Name of an instance type
	Name types.String `tfsdk:"name"`

	// PriceCentsPerHour Price of the instance type, in US cents per hour
	PriceCentsPerHour types.Int64 `tfsdk:"price_cents_per_hour"`

	// Specs Hardware configuration of an instance type
	Specs InstanceSpecsModel `tfsdk:"specs"`
}

// InstanceDataSourceModel Virtual machine (VM) in Lambda Cloud
type InstanceDataSourceModel struct {
	// FileSystemNames Names of the file systems, if any, attached to the instance
	FileSystemNames []types.String `tfsdk:"filesystem_names"`

	// Hostname assigned to this instance, which resolves to the instance's IP.
	Hostname types.String `tfsdk:"hostname"`

	// ID Unique identifier (ID) of an instance
	ID types.String `tfsdk:"id"`

	// InstanceType Hardware configuration and pricing of an instance type
	InstanceType *InstanceTypeModel `tfsdk:"instance_type"`

	// Ip IPv4 address of the instance
	Ip types.String `tfsdk:"ip"`

	// JupyterToken Secret token used to log into the jupyter lab server hosted on the instance.
	JupyterToken types.String `tfsdk:"jupyter_token"`

	// JupyterUrl URL that opens a jupyter lab notebook on the instance.
	JupyterUrl types.String `tfsdk:"jupyter_url"`

	// Name User-provided name of the instance
	Name types.String `tfsdk:"name"`

	// Region Name of the region where the instance is located
	Region *RegionModel `tfsdk:"region"`

	// SshKeyNames Names of the SSH keys allowed to access the instance
	SshKeyNames []types.String `tfsdk:"ssh_key_names"`

	// Status The current status of the instance
	Status types.String `tfsdk:"status"`
}

// InstancesDataSourceModel maps the data source schema data.
type InstancesDataSourceModel struct {
	Instances []InstanceDataSourceModel `tfsdk:"instances"`
}

// InstanceResourceModel defines parameters for provisioning an instance.
type InstanceResourceModel struct {
	// ID Unique identifier (ID) of an instance (only valid when quantity is 1)
	ID types.String `tfsdk:"id"`
	// IDs Unique identifiers (IDs) of instances
	//IDs types.Set `tfsdk:"ids"`
	// FileSystemNames Names of the file systems, if any, attached to the instance
	FileSystemNames []types.String `tfsdk:"filesystem_names"`
	// InstanceTypeName Name of an instance type
	InstanceTypeName types.String `tfsdk:"instance_type"`
	// Name User-provided name of the instance
	Name types.String `tfsdk:"name"`
	// Quantity Number of instances to provision
	//Quantity types.Int64 `tfsdk:"quantity"`
	// RegionName Name of the region where the instance is located
	RegionName types.String `tfsdk:"region"`
	// SshKeyNames Names of the SSH keys allowed to access the instance. Currently, exactly one SSH key must be specified.
	SshKeyNames []types.String `tfsdk:"ssh_key_names"`
}
