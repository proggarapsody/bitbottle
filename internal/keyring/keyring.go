package keyring

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

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

// IsHeadless returns true when no graphical session is available (CI, SSH, containers).
// Used to decide keyring timeout and whether to skip keyring entirely.
func IsHeadless() bool {
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("DOCKER") != "" {
		return true
	}
	if runtime.GOOS == "linux" && os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		return true
	}
	_, hasSSH := os.LookupEnv("SSH_TTY")
	_, hasDisplay := os.LookupEnv("DISPLAY")
	_, hasWayland := os.LookupEnv("WAYLAND_DISPLAY")
	return hasSSH || (!hasDisplay && !hasWayland)
}

func keyringTimeout() time.Duration {
	if IsHeadless() {
		return 3 * time.Second
	}
	return 60 * time.Second
}

// New returns the appropriate Keyring implementation.
// When BITBOTTLE_ALLOW_INSECURE_STORE=1 is set, a file-backed AES-256-GCM
// keyring is used instead of the OS keyring.
func New() Keyring {
	if os.Getenv("BITBOTTLE_ALLOW_INSECURE_STORE") == "1" {
		return &FileKeyring{}
	}
	return &OSKeyring{}
}

// OSKeyring delegates to the real OS keyring with a configurable timeout so
// that blocking daemon calls (libsecret, GNOME Wallet) do not hang the CLI.
//   - macOS  → Keychain
//   - Linux  → libsecret / GNOME Keyring / KDE Wallet
//   - Windows → Credential Manager
type OSKeyring struct{}

// Get retrieves a password from the OS keyring.
func (k *OSKeyring) Get(service, user string) (string, error) {
	type result struct {
		val string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		pw, err := gokeyring.Get(service, user)
		if errors.Is(err, gokeyring.ErrNotFound) {
			err = ErrNotFound
		}
		ch <- result{pw, err}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), keyringTimeout())
	defer cancel()
	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		return "", fmt.Errorf("keyring: Get timed out after %s", keyringTimeout())
	}
}

// Set stores a password in the OS keyring.
func (k *OSKeyring) Set(service, user, password string) error {
	ch := make(chan error, 1)
	go func() {
		ch <- gokeyring.Set(service, user, password)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), keyringTimeout())
	defer cancel()
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return fmt.Errorf("keyring: Set timed out after %s", keyringTimeout())
	}
}

// Delete removes a password from the OS keyring.
func (k *OSKeyring) Delete(service, user string) error {
	ch := make(chan error, 1)
	go func() {
		err := gokeyring.Delete(service, user)
		if errors.Is(err, gokeyring.ErrNotFound) {
			err = nil // idempotent
		}
		ch <- err
	}()
	ctx, cancel := context.WithTimeout(context.Background(), keyringTimeout())
	defer cancel()
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return fmt.Errorf("keyring: Delete timed out after %s", keyringTimeout())
	}
}

// ---------------------------------------------------------------------------
// FileKeyring — AES-256-GCM encrypted file-based fallback.
// Only used when BITBOTTLE_ALLOW_INSECURE_STORE=1 is set.
// Files are stored at ~/.config/bitbottle/tokens/<service>-<user>.enc
// ---------------------------------------------------------------------------

// FileKeyring stores tokens as AES-256-GCM encrypted files.
type FileKeyring struct{}

func fileKeyringDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "bitbottle", "tokens")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func fileKeyringPath(dir, service, user string) string {
	return filepath.Join(dir, fmt.Sprintf("%s-%s.enc", sanitise(service), sanitise(user)))
}

// sanitise replaces filesystem-unsafe characters with underscores.
func sanitise(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			out[i] = c
		} else {
			out[i] = '_'
		}
	}
	return string(out)
}

// fileKey derives a 32-byte AES key from a fixed machine-local secret.
// For the fallback store we use the hostname as a cheap differentiator; a
// proper KDF (HKDF/scrypt) would be better, but this keeps the implementation
// small and the threat model is "accidental plaintext exposure", not
// a determined local attacker with filesystem read access.
func fileKey() ([]byte, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "bitbottle"
	}
	raw := []byte("bitbottle-insecure-v1:" + hostname)
	// Pad/truncate to exactly 32 bytes.
	key := make([]byte, 32)
	copy(key, raw)
	return key, nil
}

// Get reads and decrypts a stored token.
func (f *FileKeyring) Get(service, user string) (string, error) {
	dir, err := fileKeyringDir()
	if err != nil {
		return "", err
	}
	path := fileKeyringPath(dir, service, user)
	ciphertext, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	key, err := fileKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("keyring: file %s is too short to be valid", path)
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("keyring: decryption failed: %w", err)
	}
	return string(plaintext), nil
}

// Set encrypts and stores a token.
func (f *FileKeyring) Set(service, user, password string) error {
	dir, err := fileKeyringDir()
	if err != nil {
		return err
	}
	key, err := fileKey()
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	path := fileKeyringPath(dir, service, user)
	return os.WriteFile(path, ciphertext, 0o600)
}

// Delete removes an encrypted token file.
func (f *FileKeyring) Delete(service, user string) error {
	dir, err := fileKeyringDir()
	if err != nil {
		return err
	}
	path := fileKeyringPath(dir, service, user)
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil // idempotent
	}
	return err
}
