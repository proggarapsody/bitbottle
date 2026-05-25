package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// ── SSH key pair ──────────────────────────────────────────────────────────────

func (h *handlers) resolveSSHKeyPairClient(req mcplib.CallToolRequest) (backend.PipelineSSHKeyPairClient, string, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", "", err
	}
	sc, err := backend.AsPipelineSSHKeyPairClient(client, hostname)
	if err != nil {
		return nil, "", "", "", err
	}
	return sc, hostname, project, slug, nil
}

func (h *handlers) viewPipelineSSHKeyPair(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sc, _, project, slug, err := h.resolveSSHKeyPairClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	kp, err := sc.GetPipelineSSHKeyPair(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(kp)
}

func (h *handlers) regeneratePipelineSSHKeyPair(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sc, _, project, slug, err := h.resolveSSHKeyPairClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	bits := req.GetInt("bits", 0)
	kp, err := sc.RegeneratePipelineSSHKeyPair(project, slug, bits)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(kp)
}

// ── Known hosts ───────────────────────────────────────────────────────────────

func (h *handlers) resolveKnownHostsClient(req mcplib.CallToolRequest) (backend.PipelineKnownHostsClient, string, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", "", err
	}
	kc, err := backend.AsPipelineKnownHostsClient(client, hostname)
	if err != nil {
		return nil, "", "", "", err
	}
	return kc, hostname, project, slug, nil
}

func (h *handlers) listPipelineKnownHosts(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	kc, _, project, slug, err := h.resolveKnownHostsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	hosts, err := kc.ListPipelineKnownHosts(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hosts)
}

func (h *handlers) viewPipelineKnownHost(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	kc, _, project, slug, err := h.resolveKnownHostsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	host, err := kc.GetPipelineKnownHost(project, slug, uuid)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(host)
}

func (h *handlers) addPipelineKnownHost(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	kc, _, project, slug, err := h.resolveKnownHostsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	hostnameArg, err := requireString(req, "hostname_arg")
	if err != nil {
		return errResultErr(err), nil
	}
	keyData := req.GetString("key", "")
	keyType := req.GetString("key_type", "RSA")
	in := backend.PipelineKnownHostInput{
		Hostname: hostnameArg,
		PublicKey: backend.PipelineSSHPublicKey{
			KeyType: keyType,
			Key:     keyData,
		},
	}
	host, err := kc.AddPipelineKnownHost(project, slug, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(host)
}

func (h *handlers) deletePipelineKnownHost(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	kc, _, project, slug, err := h.resolveKnownHostsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	if err := kc.DeletePipelineKnownHost(project, slug, uuid); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "deleted", "uuid": uuid})
}
