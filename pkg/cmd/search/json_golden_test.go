package search_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/search/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/search"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestSearchCode_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		SearchCodeFn: func(ws, q string, limit int) ([]backend.CodeSearchHit, error) {
			return []backend.CodeSearchHit{
				{
					Repository:        "acme/widgets",
					Path:              "src/main.go",
					ContentMatchCount: 1,
					FileURL:           "https://bitbucket.org/acme/widgets/src/main.go",
					ContentMatches: []backend.ContentMatch{
						{
							Line: 7,
							Segments: []backend.SearchSegment{
								{Text: "TODO", Match: true},
							},
						},
					},
				},
			}, nil
		},
	}

	f, out, _ := newFactory(t, fake)
	cmd := search.NewCmdSearchCode(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"TODO", "--workspace", "acme", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/search-code", out.String())
}
