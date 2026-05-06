package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/pkg/cmd/webhook/shared"
)

func TestParseEvents_Empty(t *testing.T) {
	t.Parallel()
	assert.Nil(t, shared.ParseEvents(""))
}

func TestParseEvents_TrimsAndDedupes(t *testing.T) {
	t.Parallel()
	got := shared.ParseEvents(" repo:push , pullrequest:created ,repo:push,, pullrequest:created ")
	assert.Equal(t, []string{"repo:push", "pullrequest:created"}, got)
}

func TestParseEvents_AllWhitespaceReturnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, shared.ParseEvents("   ,  ,  "))
}

func TestParseEvents_PreservesOrder(t *testing.T) {
	t.Parallel()
	got := shared.ParseEvents("c,a,b")
	assert.Equal(t, []string{"c", "a", "b"}, got)
}
