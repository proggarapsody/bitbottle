package cloud_test

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestCloudClient_AutoMergeUnavailableWithoutHTTP(t *testing.T) {
	for _, operation := range []string{"enable", "disable"} {
		t.Run(operation, func(t *testing.T) {
			var requests atomic.Int32
			client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				w.WriteHeader(http.StatusNoContent)
			})
			var err error
			if operation == "enable" {
				err = client.EnableAutoMerge("workspace", "repo", 42, "squash")
			} else {
				err = client.DisableAutoMerge("workspace", "repo", 42)
			}
			assert.Zero(t, requests.Load(), "unavailable Cloud auto-merge must not contact Bitbucket")
			require.ErrorIs(t, err, backend.ErrUnsupportedOnHost)
			var de *backend.DomainError
			require.ErrorAs(t, err, &de)
			assert.Equal(t, backend.ErrorCode("pr.automerge.unavailable"), de.Code)
			assert.Equal(t, "42", de.ID)
			assert.Zero(t, de.HTTPStatus())
		})
	}
}
