package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestMCP_ListServerProjects(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListServerProjectsFn: func(filter string, limit int) ([]backend.ServerProject, error) {
			return []backend.ServerProject{
				{Key: "PRJ", Name: "My Project"},
				{Key: "DEV", Name: "Dev Project"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com"})
	result, err := h.listServerProjects(context.Background(), req)
	require.NoError(t, err)
	assertJSONContains(t, result, "PRJ", "My Project")
}

func TestMCP_ListServerProjects_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// FakeClient implements ServerProjectClient, so simulate Cloud by using AsServerProjectClient
	// with a type that doesn't implement it. For MCP tests, we test the host.unsupported flow
	// by verifying the error is returned when AsServerProjectClient fails.
	// The handler passes the error through errResultErr.
	fake := &testhelpers.FakeClient{
		T: t,
		ListServerProjectsFn: func(filter string, limit int) ([]backend.ServerProject, error) {
			return nil, &backend.DomainError{
				Kind:    backend.ErrUnsupportedOnHost,
				Code:    backend.CodeHostUnsupported,
				Message: "server project management is not supported on Bitbucket Cloud",
			}
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com"})
	result, err := h.listServerProjects(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestMCP_GetServerProject(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetServerProjectFn: func(key string) (backend.ServerProject, error) {
			return backend.ServerProject{Key: key, Name: "My Project", Description: "desc"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com", "key": "PRJ"})
	result, err := h.getServerProject(context.Background(), req)
	require.NoError(t, err)
	assertJSONContains(t, result, "PRJ", "My Project")
}

func TestMCP_GetServerProject_MissingKey(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com"})
	result, err := h.getServerProject(context.Background(), req)
	require.NoError(t, err)
	assertErrorResult(t, result, "key")
}

func TestMCP_CreateServerProject(t *testing.T) {
	t.Parallel()
	var gotIn backend.CreateServerProjectInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateServerProjectFn: func(in backend.CreateServerProjectInput) (backend.ServerProject, error) {
			gotIn = in
			return backend.ServerProject{Key: in.Key, Name: in.Name}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	req := makeReq(map[string]any{
		"hostname":    "git.example.com",
		"key":         "NEWPRJ",
		"name":        "New Project",
		"description": "A project",
		"public":      false,
	})
	result, err := h.createServerProject(context.Background(), req)
	require.NoError(t, err)
	assertJSONContains(t, result, "NEWPRJ", "New Project")
	assert.Equal(t, "NEWPRJ", gotIn.Key)
	assert.Equal(t, "New Project", gotIn.Name)
}

func TestMCP_UpdateServerProject(t *testing.T) {
	t.Parallel()
	var gotKey string
	var gotIn backend.UpdateServerProjectInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateServerProjectFn: func(key string, in backend.UpdateServerProjectInput) (backend.ServerProject, error) {
			gotKey = key
			gotIn = in
			return backend.ServerProject{Key: key, Name: "Updated"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	req := makeReq(map[string]any{
		"hostname": "git.example.com",
		"key":      "PRJ",
		"name":     "Updated",
	})
	result, err := h.updateServerProject(context.Background(), req)
	require.NoError(t, err)
	assertJSONContains(t, result, "PRJ", "Updated")
	assert.Equal(t, "PRJ", gotKey)
	require.NotNil(t, gotIn.Name)
	assert.Equal(t, "Updated", *gotIn.Name)
}

func TestMCP_DeleteServerProject(t *testing.T) {
	t.Parallel()
	var gotKey string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteServerProjectFn: func(key string) error {
			gotKey = key
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com", "key": "PRJ"})
	result, err := h.deleteServerProject(context.Background(), req)
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "PRJ")
	assert.Equal(t, "PRJ", gotKey)
}
