package tag_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/tag/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/tag"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestTagList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListTagsFn: func(ns, slug string, limit int) ([]backend.Tag, error) {
			return []backend.Tag{
				{
					Name:    "v1.0.0",
					Hash:    "abc1234def567890",
					Message: "Release v1.0.0",
				},
			}, nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: tagConfig})
	factorytest.UseBackend(f, fake)
	cmd := tag.NewCmdTagList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/tag-list", out.String())
}
