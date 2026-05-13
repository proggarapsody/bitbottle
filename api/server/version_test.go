package server_test

import (
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// --- parseServerVersion tests (exercised via GetServerVersion + mock HTTP) ---

func TestParseServerVersion_Normal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  backend.ServerVersion
	}{
		{"8.5.0", backend.ServerVersion{Major: 8, Minor: 5, Patch: 0, Raw: "8.5.0"}},
		{"7.21.3", backend.ServerVersion{Major: 7, Minor: 21, Patch: 3, Raw: "7.21.3"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := versionFromResponse(t, tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseServerVersion_EmptyString(t *testing.T) {
	t.Parallel()
	got := versionFromResponse(t, "")
	assert.Equal(t, 0, got.Major)
	assert.Equal(t, 0, got.Minor)
	assert.Equal(t, 0, got.Patch)
}

func TestParseServerVersion_NotAVersion(t *testing.T) {
	t.Parallel()
	got := versionFromResponse(t, "not-a-version")
	assert.Equal(t, 0, got.Major)
	assert.Equal(t, 0, got.Minor)
	assert.Equal(t, 0, got.Patch)
	assert.Equal(t, "not-a-version", got.Raw)
}

// versionFromResponse creates a test server client backed by a handler that
// returns the given version string in the application-properties shape.
func versionFromResponse(t *testing.T, version string) backend.ServerVersion {
	t.Helper()
	body := `{"version":"` + version + `"}`
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	})
	return client.GetServerVersion()
}

// --- AtLeast table tests ---

func TestServerVersion_AtLeast(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		version backend.ServerVersion
		major   int
		minor   int
		want    bool
	}{
		{"exact match 7.2", backend.ServerVersion{Major: 7, Minor: 2}, 7, 2, true},
		{"higher minor 7.3", backend.ServerVersion{Major: 7, Minor: 3}, 7, 2, true},
		{"higher major 8.0", backend.ServerVersion{Major: 8, Minor: 0}, 7, 2, true},
		{"lower minor 7.1", backend.ServerVersion{Major: 7, Minor: 1}, 7, 2, false},
		{"lower major 6.9", backend.ServerVersion{Major: 6, Minor: 9}, 7, 2, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.version.AtLeast(tc.major, tc.minor))
		})
	}
}

// --- GetServerVersion caching test ---

func TestGetServerVersion_CachedAfterFirstCall(t *testing.T) {
	t.Parallel()
	var callCount atomic.Int32
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"8.5.0"}`)
	})

	v1 := client.GetServerVersion()
	v2 := client.GetServerVersion()

	require.Equal(t, 1, int(callCount.Load()), "GetServerVersion must issue exactly one HTTP request")
	assert.Equal(t, v1, v2)
	assert.Equal(t, 8, v1.Major)
	assert.Equal(t, 5, v1.Minor)
}

// --- AsVersionedServer ---

func TestAsVersionedServer_ServerClientImplements(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"8.5.0"}`)
	})

	vs, err := backend.AsVersionedServer(client, "https://example.com")
	require.NoError(t, err)
	require.NotNil(t, vs)
}

func TestAsVersionedServer_UnsupportedReturnsError(t *testing.T) {
	t.Parallel()
	// FakeClient does not implement VersionedServer.
	fake := &testhelpers.FakeClient{T: t}
	_, err := backend.AsVersionedServer(fake, "https://cloud.example.com")
	require.Error(t, err)
	assert.ErrorIs(t, err, backend.ErrUnsupportedOnHost)
}
