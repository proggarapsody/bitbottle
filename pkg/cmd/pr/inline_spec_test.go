package pr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestParseInlineSpec(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		spec    string
		side    string
		want    *backend.PRCommentInline
		wantErr string
	}{
		{
			name: "single-line new-side",
			spec: "main.go:42",
			side: "new",
			want: &backend.PRCommentInline{Path: "main.go", Side: "new", Line: 42},
		},
		{
			name: "single-line old-side",
			spec: "src/foo.go:7",
			side: "old",
			want: &backend.PRCommentInline{Path: "src/foo.go", Side: "old", Line: 7},
		},
		{
			name: "multi-line range",
			spec: "pkg/bar.go:10-15",
			side: "new",
			want: &backend.PRCommentInline{Path: "pkg/bar.go", Side: "new", Line: 15, StartLine: 10},
		},
		{
			name: "windows-style path with colon allowed via final colon",
			spec: "deeply/nested/path/file.go:88",
			side: "new",
			want: &backend.PRCommentInline{Path: "deeply/nested/path/file.go", Side: "new", Line: 88},
		},
		{
			name:    "empty path",
			spec:    ":42",
			side:    "new",
			wantErr: "path is required",
		},
		{
			name:    "missing line",
			spec:    "main.go:",
			side:    "new",
			wantErr: "line is required",
		},
		{
			name:    "no colon",
			spec:    "main.go",
			side:    "new",
			wantErr: "expected path:line",
		},
		{
			name:    "non-numeric line",
			spec:    "main.go:abc",
			side:    "new",
			wantErr: "line must be a positive integer",
		},
		{
			name:    "zero line",
			spec:    "main.go:0",
			side:    "new",
			wantErr: "line must be a positive integer",
		},
		{
			name:    "negative range",
			spec:    "main.go:15-10",
			side:    "new",
			wantErr: "start line",
		},
		{
			name:    "non-numeric range start",
			spec:    "main.go:a-10",
			side:    "new",
			wantErr: "line must be a positive integer",
		},
		{
			name:    "invalid side",
			spec:    "main.go:42",
			side:    "left",
			wantErr: "--side",
		},
		{
			name: "side defaults to new when empty",
			spec: "main.go:42",
			side: "",
			want: &backend.PRCommentInline{Path: "main.go", Side: "new", Line: 42},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseInlineSpec(tc.spec, tc.side)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
