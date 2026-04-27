package keyring

import (
	"errors"

	gokeyring "github.com/zalando/go-keyring"
)

// ErrNotFound is returned by Get when no credential is stored for the given
// service/user combination. Callers that treat keyring as best-effort should
// handle this alongside other errors.
var ErrNotFound = errors.New("keyring: not found")

// Keyring abstracts OS credential storage.
type Keyring interface {
	Get(service, user string) (string, error)
	Set(service, user, password string) error
	Delete(service, user string) error
}

// OSKeyring delegates to the real OS keyring:
//   - macOS  → Keychain
//   - Linux  → libsecret / GNOME Keyring / KDE Wallet
//   - Windows → Credential Manager
type OSKeyring struct{}

// Get retrieves a password from the OS keyring.
func (k *OSKeyring) Get(service, user string) (string, error) {
	pw, err := gokeyring.Get(service, user)
	if errors.Is(err, gokeyring.ErrNotFound) {
		return "", ErrNotFound
	}
	return pw, err
}

// Set stores a password in the OS keyring.
func (k *OSKeyring) Set(service, user, password string) error {
	return gokeyring.Set(service, user, password)
}

// Delete removes a password from the OS keyring.
func (k *OSKeyring) Delete(service, user string) error {
	err := gokeyring.Delete(service, user)
	if errors.Is(err, gokeyring.ErrNotFound) {
		return nil // idempotent
	}
	return err
}
