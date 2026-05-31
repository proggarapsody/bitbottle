package repoarg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

func TestParseRef(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		in      string
		want    bbrepo.RepoRef
		wantErr bool
	}{
		{
			name: "two-part PROJECT/REPO",
			in:   "MYPROJ/svc",
			want: bbrepo.RepoRef{Project: "MYPROJ", Slug: "svc"},
		},
		{
			name: "three-part HOST/PROJECT/REPO",
			in:   "bitbucket.org/ws/repo",
			want: bbrepo.RepoRef{Host: "bitbucket.org", Project: "ws", Slug: "repo"},
		},
		{
			name: "three-part server host",
			in:   "git.example.com/MYPROJ/svc",
			want: bbrepo.RepoRef{Host: "git.example.com", Project: "MYPROJ", Slug: "svc"},
		},
		{
			name:    "single bare token is not a valid ref",
			in:      "justaname",
			wantErr: true,
		},
		{
			name:    "four parts rejected",
			in:      "a/b/c/d",
			wantErr: true,
		},
		{
			name:    "empty rejected",
			in:      "",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := repoarg.ParseRef(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
