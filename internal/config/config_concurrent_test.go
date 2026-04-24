package config_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/config"
)

// TestConfig_ConcurrentSetGet verifies that concurrent Set and Get operations
// do not race (detectable with -race).
func TestConfig_ConcurrentSetGet(t *testing.T) {
	t.Parallel()

	c := config.New(t.TempDir())
	var wg sync.WaitGroup

	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set("host.example.com", config.HostConfig{
				User:        "user",
				OAuthToken:  "tok",
				GitProtocol: "ssh",
			})
		}()
	}

	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Get("host.example.com")
		}()
	}

	wg.Wait()
	hc, ok := c.Get("host.example.com")
	assert.True(t, ok)
	assert.Equal(t, "tok", hc.OAuthToken)
}

// TestConfig_ConcurrentHosts verifies Hosts() is safe under concurrent mutation.
func TestConfig_ConcurrentHosts(t *testing.T) {
	t.Parallel()

	c := config.New(t.TempDir())
	c.Set("a.example.com", config.HostConfig{User: "a", GitProtocol: "ssh"})

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.Set("b.example.com", config.HostConfig{User: "b", GitProtocol: "https"})
		}()
		go func() {
			defer wg.Done()
			_ = c.Hosts()
		}()
	}
	wg.Wait()
}

// TestConfig_ConcurrentSave verifies that concurrent Save calls do not corrupt
// the config file (last write wins, but no partial writes).
func TestConfig_ConcurrentSave(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := config.New(dir)
	c.Set("host.example.com", config.HostConfig{User: "u", OAuthToken: "tok", GitProtocol: "ssh"})

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Save()
		}()
	}
	wg.Wait()

	// File must be loadable and uncorrupted after concurrent saves.
	c2 := config.New(dir)
	require.NoError(t, c2.Load())
	hc, ok := c2.Get("host.example.com")
	assert.True(t, ok)
	assert.Equal(t, "tok", hc.OAuthToken)
}
