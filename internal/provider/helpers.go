package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// makeTfStringList converts a slice of strings to a slice of types.String.
func makeTfStringList(rawStrings []string) []types.String {
	var tfStrings = make([]types.String, 0)
	for _, str := range rawStrings {
		tfStrings = append(tfStrings, types.StringValue(str))
	}
	return tfStrings
}

// makeStringListFromTf converts a slice of types.String to a slice of strings.
func makeStringListFromTf(tfStrings []types.String) []string {
	var strings = make([]string, 0)
	for _, tfStr := range tfStrings {
		strings = append(strings, tfStr.ValueString())
	}
	return strings
}

// makeOptionalTfInt64 converts a pointer to an int to a types.Int64.
func makeOptionalTfInt64(rawInt *int) types.Int64 {
	var val int64
	if rawInt != nil {
		val = int64(*rawInt)
	}
	return types.Int64PointerValue(&val)
}
