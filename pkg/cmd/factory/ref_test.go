package factory_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// These tests pin PRD #47 host-inference behavior end-to-end through
// the public ResolveTarget surface. They previously exercised the now-
// removed Factory.ResolveRef; the rules they protect are unchanged:
// single-host = auto-pick, multi-host without flag = error, --hostname
// always wins.

func TestResolveTarget_E2E_SingleConfiguredHost_AutoPicked(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "bb.example.com:\n  oauth_token: tok\n"})
	ref, err := factory.ResolveTarget(f, []string{"MYPROJ/myrepo"}, "")
	require.NoError(t, err)
	assert.Equal(t, "bb.example.com", ref.Host)
	assert.Equal(t, "MYPROJ", ref.Project)
	assert.Equal(t, "myrepo", ref.Slug)
}

func TestResolveTarget_E2E_MultipleHosts_ErrorsWithoutFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "" +
		"bb1.example.com:\n  oauth_token: tok1\n" +
		"bb2.example.com:\n  oauth_token: tok2\n"})
	_, err := factory.ResolveTarget(f, []string{"P/r"}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple hosts")
}

func TestResolveTarget_E2E_HostnameFlag_DisambiguatesMultiHost(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "" +
		"bb1.example.com:\n  oauth_token: tok1\n" +
		"bb2.example.com:\n  oauth_token: tok2\n"})
	ref, err := factory.ResolveTarget(f, []string{"P/r"}, "bb2.example.com")
	require.NoError(t, err)
	assert.Equal(t, "bb2.example.com", ref.Host)
}
