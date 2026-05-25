package backend

import "time"

// PipelineSSHKeyPair is the domain representation of a Bitbucket Cloud
// repository-level SSH key pair used by Pipelines.
type PipelineSSHKeyPair struct {
	PublicKey    string    `json:"public_key"`
	KeyTypeLabel string    `json:"key_type"`
	Created      time.Time `json:"created"`
}

// PipelineSSHPublicKey is a public key entry in a pipeline known host.
type PipelineSSHPublicKey struct {
	KeyType string `json:"key_type"`
	Key     string `json:"key"` // base64-encoded key material
	MD5     string `json:"md5_fingerprint"`
	SHA256  string `json:"sha256_fingerprint"`
}

// PipelineKnownHost is the domain representation of a Bitbucket Cloud
// Pipelines known-host entry.
type PipelineKnownHost struct {
	UUID      string               `json:"uuid"`
	Hostname  string               `json:"hostname"`
	PublicKey PipelineSSHPublicKey `json:"public_key"`
}

// PipelineKnownHostInput carries the parameters for adding a known host.
type PipelineKnownHostInput struct {
	Hostname  string
	PublicKey PipelineSSHPublicKey
}

// PipelineSSHKeyPairClient is implemented by Cloud backends.
// Bitbucket Server/DC does not have a pipeline SSH key-pair API.
type PipelineSSHKeyPairClient interface {
	GetPipelineSSHKeyPair(ns, slug string) (PipelineSSHKeyPair, error)
	RegeneratePipelineSSHKeyPair(ns, slug string, bits int) (PipelineSSHKeyPair, error)
}

// FeaturePipelineSSHKeyPair names the pipeline-SSH-key-pair capability.
const FeaturePipelineSSHKeyPair Feature = "pipeline_ssh_key_pair"

// AsPipelineSSHKeyPairClient returns the PipelineSSHKeyPairClient view of c,
// or a typed *DomainError (Kind=ErrUnsupportedOnHost) if the backend does not
// implement the capability.
func AsPipelineSSHKeyPairClient(c Client, host string) (PipelineSSHKeyPairClient, error) {
	return requireFeature[PipelineSSHKeyPairClient](c, host, specFor(FeaturePipelineSSHKeyPair))
}

// PipelineKnownHostsClient is implemented by Cloud backends.
// Bitbucket Server/DC does not have a pipeline known-hosts API.
type PipelineKnownHostsClient interface {
	ListPipelineKnownHosts(ns, slug string) ([]PipelineKnownHost, error)
	GetPipelineKnownHost(ns, slug, uuid string) (PipelineKnownHost, error)
	AddPipelineKnownHost(ns, slug string, in PipelineKnownHostInput) (PipelineKnownHost, error)
	DeletePipelineKnownHost(ns, slug, uuid string) error
}

// FeaturePipelineKnownHosts names the pipeline-known-hosts capability.
const FeaturePipelineKnownHosts Feature = "pipeline_known_hosts"

// AsPipelineKnownHostsClient returns the PipelineKnownHostsClient view of c,
// or a typed *DomainError (Kind=ErrUnsupportedOnHost) if the backend does not
// implement the capability.
func AsPipelineKnownHostsClient(c Client, host string) (PipelineKnownHostsClient, error) {
	return requireFeature[PipelineKnownHostsClient](c, host, specFor(FeaturePipelineKnownHosts))
}
