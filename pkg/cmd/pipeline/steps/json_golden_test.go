package steps_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/pipeline/steps/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/steps"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPipelineSteps_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineStepsFn: func(ns, slug, uuid string) ([]backend.PipelineStep, error) {
			return []backend.PipelineStep{
				{
					UUID:     "step-uuid-1",
					Name:     "Build",
					State:    "SUCCESSFUL",
					Result:   "SUCCESSFUL",
					Duration: 45,
				},
			}, nil
		},
	}

	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := steps.NewCmdSteps(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "pipe-uuid-1", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/pipeline-steps", out.String())
}
