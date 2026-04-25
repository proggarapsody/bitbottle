package bbinstance_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/internal/bbinstance"
)

func TestSSHURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.SSHURL("git.example.com", "PROJ", "repo")
	assert.Equal(t, "git@git.example.com:PROJ/repo.git", got)
}

func TestHTTPSURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.HTTPSURL("git.example.com", "PROJ", "repo")
	assert.Equal(t, "https://git.example.com/scm/PROJ/repo.git", got)
}

func TestWebRepoURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.WebRepoURL("git.example.com", "PROJ", "repo")
	assert.Equal(t, "https://git.example.com/projects/PROJ/repos/repo/browse", got)
}

func TestWebPRURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.WebPRURL("git.example.com", "PROJ", "repo", 42)
	assert.Equal(t, "https://git.example.com/projects/PROJ/repos/repo/pull-requests/42", got)
}

func TestRESTBase(t *testing.T) {
	t.Parallel()
	got := bbinstance.RESTBase("git.example.com")
	assert.Equal(t, "https://git.example.com/rest/api/1.0", got)
}

func TestSupportsDraftPR_Above(t *testing.T) {
	t.Parallel()
	assert.True(t, bbinstance.SupportsDraftPR("8.9.1"))
}

func TestSupportsDraftPR_Below(t *testing.T) {
	t.Parallel()
	assert.False(t, bbinstance.SupportsDraftPR("7.0.0"))
}

func TestSupportsDraftPR_Exact(t *testing.T) {
	t.Parallel()
	assert.True(t, bbinstance.SupportsDraftPR("7.17.0"))
}

func TestIsCloud(t *testing.T) {
	t.Parallel()
	cases := []struct {
		hostname    string
		backendType string
		want        bool
	}{
		{"bitbucket.org", "", true},
		{"bitbucket.org", "server", false},
		{"bitbucket.org", "cloud", true},
		{"git.example.com", "", false},
		{"git.example.com", "cloud", true},
		{"git.example.com", "server", false},
		{"", "", false},
		{"", "cloud", true},
		{"git.example.com", "datacenter", false},
		{"bitbucket.org:443", "", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%s+%s→%v", tc.hostname, tc.backendType, tc.want), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, bbinstance.IsCloud(tc.hostname, tc.backendType))
		})
	}
}

func TestCloudSSHURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.CloudSSHURL("myworkspace", "myrepo")
	assert.Equal(t, "git@bitbucket.org:myworkspace/myrepo.git", got)
}

func TestCloudHTTPSURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.CloudHTTPSURL("myworkspace", "myrepo")
	assert.Equal(t, "https://bitbucket.org/myworkspace/myrepo.git", got)
}

func TestCloudWebRepoURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.CloudWebRepoURL("myworkspace", "myrepo")
	assert.Equal(t, "https://bitbucket.org/myworkspace/myrepo", got)
}

func TestCloudWebPRURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.CloudWebPRURL("myworkspace", "myrepo", 7)
	assert.Equal(t, "https://bitbucket.org/myworkspace/myrepo/pull-requests/7", got)
}

func TestCloudRESTBase(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://api.bitbucket.org/2.0", bbinstance.CloudRESTBase())
}

func TestIsCloud_EdgeCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		hostname    string
		backendType string
		want        bool
	}{
		{
			name:        "backendType=cloud_with_enterprise_hostname",
			hostname:    "git.internal.example.com",
			backendType: "cloud",
			want:        true,
		},
		{
			name:        "backendType=server_with_bitbucket.org",
			hostname:    "bitbucket.org",
			backendType: "server",
			want:        false,
		},
		{
			name:        "backendType_empty_with_bitbucket.org",
			hostname:    "bitbucket.org",
			backendType: "",
			want:        true,
		},
		{
			name:        "backendType_empty_with_uppercase_hostname_is_case_sensitive",
			hostname:    "BITBUCKET.ORG",
			backendType: "",
			want:        false,
		},
		{
			name:        "backendType_empty_with_api.bitbucket.org_is_false",
			hostname:    "api.bitbucket.org",
			backendType: "",
			want:        false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, bbinstance.IsCloud(tc.hostname, tc.backendType))
		})
	}
}
