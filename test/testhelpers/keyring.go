package testhelpers

import (
	"fmt"
	"sync"
)

// FakeKeyring is an in-memory test double for a system keyring.
type FakeKeyring struct {
	mu     sync.Mutex
	store  map[string]string
	GetErr error
	SetErr error
	DelErr error
}

// NewFakeKeyring returns an empty FakeKeyring.
func NewFakeKeyring() *FakeKeyring {
	return &FakeKeyring{store: map[string]string{}}
}

func key(service, user string) string {
	return service + "\x00" + user
}

// Get returns the stored password or an error. If GetErr is set it is returned.
// Missing entries return a "not found" error.
func (k *FakeKeyring) Get(service, user string) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.GetErr != nil {
		return "", k.GetErr
	}
	v, ok := k.store[key(service, user)]
	if !ok {
		return "", fmt.Errorf("secret not found in keyring: %s/%s", service, user)
	}
	return v, nil
}

// Set stores a password. If SetErr is set it is returned.
func (k *FakeKeyring) Set(service, user, password string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.SetErr != nil {
		return k.SetErr
	}
	if k.store == nil {
		k.store = map[string]string{}
	}
	k.store[key(service, user)] = password
	return nil
}

// Delete removes a stored password. If DelErr is set it is returned.
func (k *FakeKeyring) Delete(service, user string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.DelErr != nil {
		return k.DelErr
	}
	delete(k.store, key(service, user))
	return nil
}
