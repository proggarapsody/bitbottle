package server

import (
	"github.com/proggarapsody/bitbottle/api/backend"
)

// appProperties is a superset of the generated RestApplicationProperties that
// also captures the buildNumber, displayName, and type fields returned by
// GET /rest/api/1.0/application-properties.
type appProperties struct {
	Version     string `json:"version"`
	BuildNumber string `json:"buildNumber"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"` // "stash" for Server; "datacenter" for DC
}

// GetHostInfo fetches /rest/api/1.0/application-properties and returns a
// HostInfo struct populated with the backend type, version, and the list of
// Feature constants that the Server/DC adapter implements.
func (c *Client) GetHostInfo() (backend.HostInfo, error) {
	var props appProperties
	if err := c.http.GetJSON("/application-properties", &props); err != nil {
		return backend.HostInfo{}, err
	}

	backendType := "server"
	if props.Type == "datacenter" {
		backendType = "datacenter"
	}

	displayName := props.DisplayName
	if displayName == "" {
		displayName = "Bitbucket Server"
	}

	var features []string
	for _, spec := range backend.AllFeatureSpecs {
		if spec.ServerSupport {
			features = append(features, string(spec.Feature))
		}
	}

	return backend.HostInfo{
		BackendType:       backendType,
		BaseURL:           c.host,
		Version:           props.Version,
		BuildNumber:       props.BuildNumber,
		DisplayName:       displayName,
		SupportedFeatures: features,
	}, nil
}
