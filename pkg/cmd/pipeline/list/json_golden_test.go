package list_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/pipeline/list/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPipelineList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{
				{
					UUID:        "pipe-uuid-1",
					BuildNumber: 99,
					State:       "SUCCESSFUL",
					RefName:     "main",
					Duration:    120,
					WebURL:      "https://bitbucket.org/myworkspace/my-service/addon/pipelines/home#!/results/99",
				},
			}, nil
		},
	}

	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/pipeline-list", out.String())
}
