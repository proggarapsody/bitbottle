package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListPipelineSchedules_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListPipelineSchedules("myworkspace", "my-service")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pipelines_config/schedules", gotPath)
}

func TestCloudClient_ListPipelineSchedules_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{sched-1}","enabled":true,"cron_expression":"0 0 * * *","target":{"branch":"main"}},
			{"uuid":"{sched-2}","enabled":false,"cron_expression":"0 12 * * 1","target":{"branch":"develop"}}
		]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	schedules, err := client.ListPipelineSchedules("myworkspace", "my-service")
	require.NoError(t, err)
	require.Len(t, schedules, 2)
	assert.Equal(t, "sched-1", schedules[0].UUID)
	assert.True(t, schedules[0].Enabled)
	assert.Equal(t, "0 0 * * *", schedules[0].CronExpression)
	assert.Equal(t, "main", schedules[0].Branch)
	assert.Equal(t, "sched-2", schedules[1].UUID)
	assert.False(t, schedules[1].Enabled)
	assert.Equal(t, "develop", schedules[1].Branch)
}

func TestCloudClient_ListPipelineSchedules_UUIDHasNoBraces(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{sched-abc}","enabled":true,"cron_expression":"0 0 * * *","target":{"branch":"main"}}
		]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	schedules, err := client.ListPipelineSchedules("myworkspace", "my-service")
	require.NoError(t, err)
	require.Len(t, schedules, 1)
	assert.NotContains(t, schedules[0].UUID, "{")
	assert.NotContains(t, schedules[0].UUID, "}")
}

func TestCloudClient_CreatePipelineSchedule_PostsCorrectBody(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{new-sched}","enabled":true,"cron_expression":"0 0 * * *","target":{"ref_name":"main","ref_type":"branch","type":"pipeline_ref_target"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	got, err := client.CreatePipelineSchedule("myworkspace", "my-service", backend.PipelineScheduleInput{
		CronExpression: "0 0 * * *",
		Branch:         "main",
		Enabled:        true,
	})
	require.NoError(t, err)
	assert.Equal(t, "new-sched", got.UUID)
	assert.True(t, got.Enabled)
	target, _ := gotBody["target"].(map[string]any)
	require.NotNil(t, target)
	assert.Equal(t, "branch", target["ref_type"])
	assert.Equal(t, "main", target["ref_name"])
	assert.Equal(t, "pipeline_ref_target", target["type"])
	assert.Equal(t, "0 0 * * *", gotBody["cron_expression"])
	assert.Equal(t, true, gotBody["enabled"])
}

func TestCloudClient_CreatePipelineSchedule_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{x}","enabled":true,"cron_expression":"0 0 * * *","target":{"branch":"main"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.CreatePipelineSchedule("myworkspace", "my-service", backend.PipelineScheduleInput{
		CronExpression: "0 0 * * *",
		Branch:         "main",
		Enabled:        true,
	})
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pipelines_config/schedules", gotPath)
}

func TestCloudClient_DeletePipelineSchedule_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	var gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	err := client.DeletePipelineSchedule("myworkspace", "my-service", "sched-abc")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/myworkspace/my-service/pipelines_config/schedules/{sched-abc}", gotPath)
}
