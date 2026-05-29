package cloud

import (
	"github.com/proggarapsody/bitbottle/api/backend"
)

// GetHostInfo returns static metadata about the Bitbucket Cloud backend.
// No HTTP request is made: the backend type is fixed, the feature list is
// derived from AllFeatureSpecs at compile time, and the version is omitted
// (Cloud is always "rolling").
func (c *Client) GetHostInfo() (backend.HostInfo, error) {
	var features []string
	for _, spec := range backend.AllFeatureSpecs {
		if spec.CloudSupport {
			features = append(features, string(spec.Feature))
		}
	}

	return backend.HostInfo{
		BackendType:       "cloud",
		BaseURL:           c.baseURL,
		DisplayName:       "Bitbucket Cloud",
		SupportedFeatures: features,
	}, nil
}
