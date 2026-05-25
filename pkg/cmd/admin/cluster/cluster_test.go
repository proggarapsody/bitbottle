package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/cluster"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCluster_SingleNode(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetClusterNodesFn: func() ([]backend.ClusterNode, error) {
			return []backend.ClusterNode{
				{NodeId: "node-1", Name: "Primary", Address: "10.0.0.1", State: "ACTIVE", Local: true},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cluster.NewCmdCluster(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "node-1")
	assert.Contains(t, out.String(), "Primary")
	assert.Contains(t, out.String(), "10.0.0.1")
	assert.Contains(t, out.String(), "ACTIVE")
}

func TestCluster_MultipleNodes(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetClusterNodesFn: func() ([]backend.ClusterNode, error) {
			return []backend.ClusterNode{
				{NodeId: "node-1", Name: "Primary", Address: "10.0.0.1", State: "ACTIVE", Local: true},
				{NodeId: "node-2", Name: "Secondary", Address: "10.0.0.2", State: "ACTIVE", Local: false},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cluster.NewCmdCluster(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "node-1")
	assert.Contains(t, out.String(), "node-2")
	assert.Contains(t, out.String(), "Secondary")
}

func TestCluster_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetClusterNodesFn: func() ([]backend.ClusterNode, error) {
			return []backend.ClusterNode{
				{NodeId: "node-1", Name: "Primary", Address: "10.0.0.1", State: "ACTIVE", Local: true},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cluster.NewCmdCluster(f, nil)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"NodeId":"node-1"`)
	assert.Contains(t, out.String(), `"Local":true`)
}

func TestCluster_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := cluster.NewCmdCluster(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
