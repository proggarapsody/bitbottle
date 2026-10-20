package cloud_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestCloudClient_MergePRConfirmsCompletion(t *testing.T) {
	for _, tc := range []struct {
		name, response, readback string
		status                   int
		pending, malformed       bool
	}{
		{name: "queued", response: `{"id":42,"state":"OPEN","queued":true}`, status: 200, pending: true},
		{name: "merged", response: `{"id":42,"state":"MERGED"}`, status: 200},
		{name: "empty accepted still open", status: 202, readback: `{"id":42,"state":"OPEN"}`, pending: true},
		{name: "empty accepted confirmed merged", status: 202, readback: `{"id":42,"state":"MERGED"}`},
		{name: "truncated response", status: 202, response: `{"state":`, malformed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var methods []string
			client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
				methods = append(methods, r.Method)
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					_, _ = io.WriteString(w, tc.readback)
					return
				}
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.response)
			})
			pr, err := client.MergePR("workspace", "repo", 42, backend.MergePRInput{})
			if tc.pending {
				var de *backend.DomainError
				require.ErrorAs(t, err, &de)
				assert.Equal(t, backend.ErrorCode("pr.merge.pending"), de.Code)
			} else if tc.malformed {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "MERGED", pr.State)
			}
			if tc.readback != "" {
				assert.Equal(t, []string{"POST", "GET"}, methods)
			} else {
				assert.Equal(t, []string{"POST"}, methods)
			}
		})
	}
}
