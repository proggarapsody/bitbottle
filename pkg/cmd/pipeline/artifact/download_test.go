package artifact_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/artifact"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdDownload_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, nil)
	assert.NotNil(t, cmd.Flag("step"))
	assert.NotNil(t, cmd.Flag("name"))
	assert.NotNil(t, cmd.Flag("out"))
}

func TestNewCmdDownload_RequiresPipelineUUID(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, nil)
	cmd.SetArgs([]string{"--step", "s", "--name", "f.zip"})
	require.Error(t, cmd.Execute())
}

func TestNewCmdDownload_RequiresStep(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--name", "f.zip"})
	require.Error(t, cmd.Execute())
}

func TestNewCmdDownload_RequiresName(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "s"})
	require.Error(t, cmd.Execute())
}

func TestDownload_StdoutMode_WritesToOut(t *testing.T) {
	t.Parallel()
	content := "artifact-binary-data"
	fake := &testhelpers.FakeClient{
		T: t,
		DownloadPipelineArtifactFn: func(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error {
			_, err := io.WriteString(out, content)
			return err
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "step-uuid", "--name", "build.tar.gz", "--out", "-"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, content, out.String())
}

func TestDownload_StdoutMode_NoStderrMessage(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DownloadPipelineArtifactFn: func(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error {
			_, err := io.WriteString(out, "data")
			return err
		},
	}
	f, _, errOut := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "step-uuid", "--name", "f.zip", "--out", "-"})
	require.NoError(t, cmd.Execute())
	assert.Empty(t, errOut.String())
}

func TestDownload_NotCloudCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	type serverOnlyFake struct{ backend.Client }
	fake := &serverOnlyFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "step-uuid", "--name", "f.zip", "--out", "-"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline artifacts")
}

func TestDownload_PassesCorrectParams(t *testing.T) {
	t.Parallel()
	var gotWS, gotSlug, gotPipeUUID, gotStepUUID, gotName string
	fake := &testhelpers.FakeClient{
		T: t,
		DownloadPipelineArtifactFn: func(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error {
			gotWS = ws
			gotSlug = slug
			gotPipeUUID = pipelineUUID
			gotStepUUID = stepUUID
			gotName = name
			_, err := io.WriteString(out, "ok")
			return err
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, nil)
	cmd.SetArgs([]string{"my-pipe", "myws/repo", "--step", "my-step", "--name", "build.zip", "--out", "-"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "myws", gotWS)
	assert.Equal(t, "repo", gotSlug)
	assert.Equal(t, "my-pipe", gotPipeUUID)
	assert.Equal(t, "my-step", gotStepUUID)
	assert.Equal(t, "build.zip", gotName)
}

func TestDownload_RunF_Override(t *testing.T) {
	t.Parallel()
	var called bool
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, func(opts *artifact.DownloadOptions) error {
		called = true
		assert.Equal(t, "my-pipe", opts.PipelineUUID)
		assert.Equal(t, "my-step", opts.StepUUID)
		assert.Equal(t, "build.zip", opts.Name)
		return nil
	})
	cmd.SetArgs([]string{"my-pipe", "myws/repo", "--step", "my-step", "--name", "build.zip"})
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
}

func TestDownload_DownloadError_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DownloadPipelineArtifactFn: func(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error {
			return &backend.DomainError{
				Kind:    backend.ErrNotFound,
				Code:    "artifact.not_found",
				Message: "artifact not found",
			}
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "step-uuid", "--name", "missing.zip", "--out", "-"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "artifact not found") || strings.Contains(err.Error(), "not found"))
}

func TestDownload_ExistingFile_WithoutClobber_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	existing := filepath.Join(dir, "build.zip")
	require.NoError(t, os.WriteFile(existing, []byte("old"), 0o644))

	fake := &testhelpers.FakeClient{
		T: t,
		DownloadPipelineArtifactFn: func(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error {
			_, err := io.WriteString(out, "new")
			return err
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, func(opts *artifact.DownloadOptions) error {
		// Override dest to use temp dir path
		opts.Out = ""
		opts.Name = existing
		opts.Clobber = false
		return artifact.RunDownload(f, opts)
	})
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "s", "--name", "build.zip"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDownload_ExistingFile_WithClobber_Succeeds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	existing := filepath.Join(dir, "build.zip")
	require.NoError(t, os.WriteFile(existing, []byte("old"), 0o644))

	fake := &testhelpers.FakeClient{
		T: t,
		DownloadPipelineArtifactFn: func(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error {
			_, err := io.WriteString(out, "new-content")
			return err
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdDownload(f, func(opts *artifact.DownloadOptions) error {
		opts.Out = ""
		opts.Name = existing
		opts.Clobber = true
		return artifact.RunDownload(f, opts)
	})
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "s", "--name", "build.zip", "--clobber"})
	require.NoError(t, cmd.Execute())
	got, err := os.ReadFile(existing)
	require.NoError(t, err)
	assert.Equal(t, "new-content", string(got))
}
