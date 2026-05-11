package watch_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/watch"
)

func TestNewCmdWatch_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := watch.NewCmdWatch(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service"}) // only 1 arg — should error
	err := cmd.Execute()
	require.Error(t, err)
}

func TestNewCmdWatch_HasIntervalFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := watch.NewCmdWatch(f, nil)
	flag := cmd.Flag("interval")
	require.NotNil(t, flag)
	assert.Equal(t, "5", flag.DefValue)
}

func TestNewCmdWatch_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := watch.NewCmdWatch(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestNewCmdWatch_RunFReceivesArgs(t *testing.T) {
	t.Parallel()

	var gotOpts *watch.Options
	runF := func(opts *watch.Options) error {
		gotOpts = opts
		return nil
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := watch.NewCmdWatch(f, runF)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc-uuid-123"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotOpts)
	assert.Equal(t, []string{"MYPROJ/my-service", "abc-uuid-123"}, gotOpts.Args)
}

func TestNewCmdWatch_RunFReceivesInterval(t *testing.T) {
	t.Parallel()

	var gotOpts *watch.Options
	runF := func(opts *watch.Options) error {
		gotOpts = opts
		return nil
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := watch.NewCmdWatch(f, runF)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc-uuid-123", "--interval", "30"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotOpts)
	assert.Equal(t, 30, gotOpts.Interval)
}

func TestNewCmdWatch_RunFErrorPropagates(t *testing.T) {
	t.Parallel()

	runF := func(_ *watch.Options) error {
		return errors.New("pipeline ended with state: FAILED")
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := watch.NewCmdWatch(f, runF)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc-uuid-123"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FAILED")
}
