package pr_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPRActivity_RendersTable(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRActivityFn: func(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
			assert.Equal(t, 42, id)
			return []backend.PRActivityEvent{
				{Type: "approval", Actor: backend.User{Slug: "alice", DisplayName: "Alice"}, CreatedAt: now},
				{Type: "comment", Actor: backend.User{Slug: "bob", DisplayName: "Bob"}, CreatedAt: now.Add(time.Hour)},
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRActivity(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "approval")
	assert.Contains(t, got, "Alice")
	assert.Contains(t, got, "comment")
	assert.Contains(t, got, "Bob")
}

func TestPRActivity_EmptyResultPrintsSentinel(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRActivityFn: func(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
			return nil, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRActivity(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "No activity found.")
}

func TestPRActivity_LimitFlagPassedThrough(t *testing.T) {
	t.Parallel()
	var gotLimit int
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRActivityFn: func(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
			gotLimit = limit
			return nil, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRActivity(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--limit", "5"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 5, gotLimit)
}

func TestPRActivity_APIError_Propagated(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRActivityFn: func(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
			return nil, errors.New("server error")
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRActivity(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server error")
}

func TestPRActivity_JSONFlag(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRActivityFn: func(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
			return []backend.PRActivityEvent{
				{Type: "approval", Actor: backend.User{Slug: "alice"}, CreatedAt: now},
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRActivity(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"type"`)
	assert.Contains(t, got, `"approval"`)
	assert.Contains(t, got, `"createdAt"`)
}
