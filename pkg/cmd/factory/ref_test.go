package factory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// TestResolveRef_SingleConfiguredHost_AutoPicked pins PRD #47 host-inference
// rule (a): when exactly one host is configured, ResolveRef picks it without
// requiring --hostname.
func TestResolveRef_SingleConfiguredHost_AutoPicked(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n",
	})
	ref, err := f.ResolveRef("MYPROJ/myrepo", "")
	require.NoError(t, err)
	assert.Equal(t, "bb.example.com", ref.Host)
	assert.Equal(t, "MYPROJ", ref.Project)
	assert.Equal(t, "myrepo", ref.Slug)
}

// TestResolveRef_MultipleHosts_ErrorsWithoutFlag pins PRD #47 host-inference
// rule: with two hosts and no --hostname, the resolver errors rather than
// guessing.
func TestResolveRef_MultipleHosts_ErrorsWithoutFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "" +
			"bb1.example.com:\n  oauth_token: tok1\n" +
			"bb2.example.com:\n  oauth_token: tok2\n",
	})
	_, err := f.ResolveRef("P/r", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple hosts")
}

// TestResolveRef_HostnameFlag_DisambiguatesMultiHost pins the rule that an
// explicit --hostname picks the host even when multiple are configured.
func TestResolveRef_HostnameFlag_DisambiguatesMultiHost(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "" +
			"bb1.example.com:\n  oauth_token: tok1\n" +
			"bb2.example.com:\n  oauth_token: tok2\n",
	})
	ref, err := f.ResolveRef("P/r", "bb2.example.com")
	require.NoError(t, err)
	assert.Equal(t, "bb2.example.com", ref.Host)
}
