package keyring

// Keyring abstracts OS credential storage.
type Keyring interface {
	Get(service, user string) (string, error)
	Set(service, user, password string) error
	Delete(service, user string) error
}

// OSKeyring wraps the real OS keyring via go-keyring.
type OSKeyring struct{}

// Get retrieves a password.
func (k *OSKeyring) Get(service, user string) (string, error) {
	panic("not implemented")
}

// Set stores a password.
func (k *OSKeyring) Set(service, user, password string) error {
	panic("not implemented")
}

// Delete removes a password.
func (k *OSKeyring) Delete(service, user string) error {
	panic("not implemented")
}
