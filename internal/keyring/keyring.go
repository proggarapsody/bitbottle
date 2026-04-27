package keyring

import "errors"

// ErrNotImplemented is returned by OSKeyring when the OS keyring is not yet wired up.
var ErrNotImplemented = errors.New("keyring: not implemented")

// Keyring abstracts OS credential storage.
type Keyring interface {
	Get(service, user string) (string, error)
	Set(service, user, password string) error
	Delete(service, user string) error
}

// OSKeyring wraps the real OS keyring via go-keyring.
// Currently a no-op stub: all operations return ErrNotImplemented so callers
// that treat keyring storage as best-effort continue to work correctly.
type OSKeyring struct{}

// Get retrieves a password.
func (k *OSKeyring) Get(service, user string) (string, error) {
	return "", ErrNotImplemented
}

// Set stores a password.
func (k *OSKeyring) Set(service, user, password string) error {
	return ErrNotImplemented
}

// Delete removes a password.
func (k *OSKeyring) Delete(service, user string) error {
	return ErrNotImplemented
}
