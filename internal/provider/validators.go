package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.List = &ListMaxLength{}
var _ defaults.Int64 = &Int64Default{}
var _ defaults.List = &ListDefaultEmpty{}

// ListMaxLength is a schema validator for the length of types.List.
type ListMaxLength struct {
	max int
}

func (m ListMaxLength) Description(ctx context.Context) string {
	return "Max length validator"
}

func (m ListMaxLength) MarkdownDescription(ctx context.Context) string {
	return "Max length validator"
}

func (m ListMaxLength) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	if request.ConfigValue.IsNull() {
		return
	}
	if len(request.ConfigValue.Elements()) > m.max {
		response.Diagnostics.AddError("List too long", fmt.Sprintf("List is too long, max length is %d", m.max))
	}
}

// Int64Default is a schema default value for types.Int64 attributes.
type Int64Default struct {
	defaultValue int64
}

func (d Int64Default) Description(ctx context.Context) string {
	return "Default Value for Int64 fields"
}

func (d Int64Default) MarkdownDescription(ctx context.Context) string {
	return "Default Value for Int64 fields"
}

func (d Int64Default) DefaultInt64(ctx context.Context, request defaults.Int64Request, response *defaults.Int64Response) {
	response.PlanValue = types.Int64Value(d.defaultValue)
}

// ListDefaultEmpty is a schema default value for types.List attributes.
type ListDefaultEmpty struct {
	ElementType attr.Type
}

func (d ListDefaultEmpty) Description(ctx context.Context) string {
	return "Default Value of empty list for List fields"
}

func (d ListDefaultEmpty) MarkdownDescription(ctx context.Context) string {
	return "Default Value of empty list for List fields"
}

func (d ListDefaultEmpty) DefaultList(ctx context.Context, request defaults.ListRequest, response *defaults.ListResponse) {
	response.PlanValue = types.ListNull(d.ElementType)
}
