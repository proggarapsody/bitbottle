package report_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/codeinsights/report"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: server\n"

func runner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(testhelpers.RunResponse{
		Stdout: "https://bitbucket.org/MYPROJ/my-service.git\n",
	})
}

func newFactory(t *testing.T, fake *testhelpers.FakeClient) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)
	r := runner()
	f.GitRunner = func() run.Runner { return r }
	return f, out, errOut
}

// noInsightsFake satisfies backend.Client without CodeInsightsClient.
type noInsightsFake struct{ backend.Client }

// ── list ──────────────────────────────────────────────────────────────────────

func TestReportList_PrintsRows(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListReportsFn: func(project, slug, hash string) ([]backend.CodeInsightsReport, error) {
			assert.Equal(t, "MYPROJ", project)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "abc123", hash)
			return []backend.CodeInsightsReport{
				{Key: "tool-1", Title: "Tool One", Result: "PASS", ReportType: "TESTING"},
			}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := report.NewCmdList(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc123"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "tool-1")
	assert.Contains(t, out.String(), "PASS")
}

func TestReportList_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListReportsFn: func(project, slug, hash string) ([]backend.CodeInsightsReport, error) {
			return []backend.CodeInsightsReport{{Key: "k1", Title: "T", Result: "FAIL"}}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := report.NewCmdList(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc123", "--json", "key,result"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"key":"k1"`)
	assert.Contains(t, out.String(), `"result":"FAIL"`)
}

func TestReportList_Unsupported(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, noInsightsFake{Client: &testhelpers.FakeClient{T: t}})
	r := runner()
	f.GitRunner = func() run.Runner { return r }

	cmd := report.NewCmdList(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc123"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Server / Data Center only")
}

// ── view ──────────────────────────────────────────────────────────────────────

func TestReportView_PrintsSingleRow(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetReportFn: func(project, slug, hash, key string) (backend.CodeInsightsReport, error) {
			assert.Equal(t, "my-key", key)
			return backend.CodeInsightsReport{Key: key, Title: "T", Result: "PASS"}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := report.NewCmdView(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc123", "my-key"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "my-key")
	assert.Contains(t, out.String(), "PASS")
}

// ── set ───────────────────────────────────────────────────────────────────────

func TestReportSet_CallsWithCorrectInput(t *testing.T) {
	t.Parallel()
	var gotIn backend.CodeInsightsReportInput
	fake := &testhelpers.FakeClient{T: t,
		SetReportFn: func(project, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error) {
			gotIn = in
			return backend.CodeInsightsReport{Key: key, Title: in.Title, Result: in.Result}, nil
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := report.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc123", "my-key",
		"--title", "Scan", "--result", "fail", "--report-type", "security"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Scan", gotIn.Title)
	assert.Equal(t, "FAIL", gotIn.Result)
	assert.Equal(t, "SECURITY", gotIn.ReportType)
}

func TestReportSet_ParsesDatum(t *testing.T) {
	t.Parallel()
	var gotIn backend.CodeInsightsReportInput
	fake := &testhelpers.FakeClient{T: t,
		SetReportFn: func(project, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error) {
			gotIn = in
			return backend.CodeInsightsReport{Key: key}, nil
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := report.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc123", "k",
		"--title", "T", "--result", "PASS",
		"--data", "Coverage=87.5:PERCENTAGE"})
	require.NoError(t, cmd.Execute())
	require.Len(t, gotIn.Data, 1)
	assert.Equal(t, "Coverage", gotIn.Data[0].Title)
	assert.Equal(t, "PERCENTAGE", gotIn.Data[0].Type)
	assert.Equal(t, "87.5", gotIn.Data[0].Value)
}

func TestReportSet_InvalidDatum(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newFactory(t, fake)
	cmd := report.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc123", "k",
		"--title", "T", "--result", "PASS",
		"--data", "bad-datum-no-type"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TITLE=VALUE:TYPE")
}

// ── delete ────────────────────────────────────────────────────────────────────

func TestReportDelete_Succeeds(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		DeleteReportFn: func(project, slug, hash, key string) error {
			called = true
			assert.Equal(t, "del-key", key)
			return nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := report.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc123", "del-key"})
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
	assert.Contains(t, out.String(), "del-key")
}
