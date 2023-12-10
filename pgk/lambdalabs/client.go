package lambdalabs

import (
	"fmt"
	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
)

// NewAuthenticatedClient creates a new Lambda Labs API client with a bearer token
func NewAuthenticatedClient(host string, apiKey string, opts ...ClientOption) (*ClientWithResponses, error) {

	if host == "" {
		host = "https://cloud.lambdalabs.com/api/v1"
	}
	if apiKey == "" {
		return nil, fmt.Errorf("missing Lambda Labs API Key")
	}
	bearerTokenProvider, bearerTokenProviderErr := securityprovider.NewSecurityProviderBearerToken(apiKey)
	if bearerTokenProviderErr != nil {
		return nil, fmt.Errorf("unable to create Lambda Labs API client: %w", bearerTokenProviderErr)
	}
	opts = append(opts, WithRequestEditorFn(bearerTokenProvider.Intercept))
	return NewClientWithResponses(host, opts...)
}
