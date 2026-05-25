package mcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listRepoDownloads(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	rc, err := backend.AsRepoDownloadClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	downloads, err := rc.ListRepoDownloads(project, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(downloads)
}

func (h *handlers) uploadRepoDownload(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	encoded, err := requireString(req, "file_content_base64")
	if err != nil {
		return errResultErr(err), nil
	}
	decoded, decErr := base64.StdEncoding.DecodeString(encoded)
	if decErr != nil {
		return errResult(fmt.Sprintf("invalid base64 in file_content_base64: %v", decErr)), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	rc, err := backend.AsRepoDownloadClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	d, err := rc.UploadRepoDownload(project, slug, name, bytes.NewReader(decoded))
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(d)
}

func (h *handlers) deleteRepoDownload(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	rc, err := backend.AsRepoDownloadClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := rc.DeleteRepoDownload(project, slug, name); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Deleted download %s from %s/%s.", name, project, slug)), nil
}
