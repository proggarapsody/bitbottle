package backend

import "fmt"

// SSHKeyClient is implemented by Cloud backends only.
// User SSH keys are a Cloud-only concept; Server/DC uses deploy keys instead.
type SSHKeyClient interface {
	ListSSHKeys() ([]SSHKey, error)
	AddSSHKey(input SSHKeyInput) (SSHKey, error)
	DeleteSSHKey(id int) error
}

// FeatureSSHKeys names the ssh-keys capability for typed-error reporting.
const FeatureSSHKeys Feature = "ssh_keys"

// AsSSHKeyClient returns the SSHKeyClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the SSHKeys capability.
func AsSSHKeyClient(c Client, host string) (SSHKeyClient, error) {
	sk, ok := c.(SSHKeyClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureSSHKeys),
			Message: fmt.Sprintf("user SSH keys are not supported on %s", host),
		}
	}
	return sk, nil
}
