package audit_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/audit"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// cloudConfig is a single-host Cloud config — audit log is Cloud only.
const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdAudit_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := audit.NewCmdAudit(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("action"))
	assert.NotNil(t, cmd.Flag("from"))
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
	assert.Equal(t, "25", cmd.Flag("limit").DefValue)
}

func TestNewCmdAudit_AcceptsOptionalWorkspaceArg(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := audit.NewCmdAudit(f, func(opts *audit.Options) error {
		gotWorkspace = opts.Workspace
		return nil
	})
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myws"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "myws", gotWorkspace)
}

func TestNewCmdAudit_RejectsTooManyArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := audit.NewCmdAudit(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"ws1", "ws2"})
	require.Error(t, cmd.Execute())
}

func TestAudit_PrintsEventFields(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListAuditLogFn: func(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error) {
			assert.Equal(t, "acme", workspace)
			return []backend.AuditEvent{
				{
					Actor:     backend.AuditActor{AccountID: "aid-1", DisplayName: "Alice", NickName: "alice"},
					Action:    "workspace.member.create",
					Object:    backend.AuditObject{Type: "team", Name: "acme-obj"},
					CreatedAt: "2024-01-15T10:00:00.000000+00:00",
				},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := audit.NewCmdAudit(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "Alice")
	assert.Contains(t, got, "workspace.member.create")
	assert.Contains(t, got, "acme-obj")
}

func TestAudit_ForwardsActionAndFromFlags(t *testing.T) {
	t.Parallel()
	var gotOpts backend.AuditLogOpts
	fake := &testhelpers.FakeClient{
		T: t,
		ListAuditLogFn: func(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error) {
			gotOpts = opts
			return nil, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := audit.NewCmdAudit(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme", "--action", "workspace.member.create", "--from", "2024-01-01T00:00:00Z", "--limit", "10"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "workspace.member.create", gotOpts.Action)
	assert.Equal(t, "2024-01-01T00:00:00Z", gotOpts.From)
	assert.Equal(t, 10, gotOpts.Limit)
}

func TestAudit_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListAuditLogFn: func(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error) {
			return []backend.AuditEvent{
				{
					Actor:     backend.AuditActor{AccountID: "aid-1", DisplayName: "Alice", NickName: "alice"},
					Action:    "workspace.member.create",
					Object:    backend.AuditObject{Type: "team", Name: "acme-obj"},
					CreatedAt: "2024-01-15T10:00:00.000000+00:00",
				},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := audit.NewCmdAudit(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "Alice")
	assert.Contains(t, got, "workspace.member.create")
	assert.Contains(t, got, "acme-obj")
}

func TestAudit_ClientError_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListAuditLogFn: func(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error) {
			return nil, errors.New("403 Forbidden")
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := audit.NewCmdAudit(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestAudit_NoWorkspaceArg_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := audit.NewCmdAudit(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

// noAuditFake wraps backend.Client without satisfying AuditClient,
// simulating a Server/DC backend invocation.
type noAuditFake struct {
	backend.Client
}

func TestAudit_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noAuditFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := audit.NewCmdAudit(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}
