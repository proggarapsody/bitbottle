package server

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// diffPathStr returns the path string from a RestDiffPath pointer.
func diffPathStr(p *servergen.RestDiffPath) string {
	if p == nil {
		return ""
	}
	if p.ToString != "" {
		return p.ToString
	}
	return p.Name
}

// buildUnifiedDiff reconstructs a unified diff text from Server's JSON representation.
func buildUnifiedDiff(diffs []servergen.RestDiff) string {
	var sb strings.Builder
	for _, d := range diffs {
		srcPath := diffPathStr(d.Source)
		dstPath := diffPathStr(d.Destination)
		if srcPath == "" && dstPath == "" {
			continue
		}
		fromFile := srcPath
		toFile := dstPath
		if fromFile == "" {
			fromFile = "/dev/null"
		}
		if toFile == "" {
			toFile = "/dev/null"
		}
		fmt.Fprintf(&sb, "--- a/%s\n", fromFile)
		fmt.Fprintf(&sb, "+++ b/%s\n", toFile)
		for _, h := range d.Hunks {
			fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n",
				h.SourceLine, h.SourceSpan,
				h.DestinationLine, h.DestinationSpan,
			)
			for _, seg := range h.Segments {
				for _, line := range seg.Lines {
					switch seg.Type {
					case "ADDED":
						fmt.Fprintf(&sb, "+%s\n", line.Line)
					case "REMOVED":
						fmt.Fprintf(&sb, "-%s\n", line.Line)
					default: // CONTEXT
						fmt.Fprintf(&sb, " %s\n", line.Line)
					}
				}
			}
		}
	}
	return sb.String()
}

// GetDiff returns a unified diff between two refs for a repository on Server/DC.
// Server endpoint: GET /rest/api/1.0/projects/{ns}/repos/{slug}/diff?since={from}&until={to}&contextLines=5
func (c *Client) GetDiff(ns, slug, from, to string) (string, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/diff", url.PathEscape(ns), url.PathEscape(slug))
	path += fmt.Sprintf("?since=%s&until=%s&contextLines=5", url.QueryEscape(from), url.QueryEscape(to))
	var resp servergen.RestDiffResponse
	if err := c.getJSON(path, &resp); err != nil {
		return "", err
	}
	return buildUnifiedDiff(resp.Diffs), nil
}

// GetDiffStat returns the diff summary between two refs for a repository on Server/DC.
// It reuses the same JSON response as GetDiff and counts files/lines from hunk segments.
func (c *Client) GetDiffStat(ns, slug, from, to string) (backend.DiffStat, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/diff", url.PathEscape(ns), url.PathEscape(slug))
	path += fmt.Sprintf("?since=%s&until=%s&contextLines=0", url.QueryEscape(from), url.QueryEscape(to))
	var resp servergen.RestDiffResponse
	if err := c.getJSON(path, &resp); err != nil {
		return backend.DiffStat{}, err
	}
	stat := backend.DiffStat{
		FilesChanged: len(resp.Diffs),
		Files:        make([]backend.DiffStatEntry, 0, len(resp.Diffs)),
	}
	for _, d := range resp.Diffs {
		entry := diffEntryFromServer(d)
		stat.Additions += entry.Additions
		stat.Deletions += entry.Deletions
		stat.Files = append(stat.Files, entry)
	}
	return stat, nil
}

// diffEntryFromServer maps a single Server diff to a DiffStatEntry.
func diffEntryFromServer(d servergen.RestDiff) backend.DiffStatEntry {
	srcPath := diffPathStr(d.Source)
	dstPath := diffPathStr(d.Destination)

	path := dstPath
	if path == "" {
		path = srcPath
	}

	var status string
	switch {
	case srcPath == "":
		status = "added"
	case dstPath == "":
		status = "deleted"
	case srcPath != dstPath:
		status = "renamed"
	default:
		status = "modified"
	}

	var adds, dels int
	for _, h := range d.Hunks {
		for _, seg := range h.Segments {
			switch seg.Type {
			case "ADDED":
				adds += len(seg.Lines)
			case "REMOVED":
				dels += len(seg.Lines)
			}
		}
	}
	return backend.DiffStatEntry{
		Path:      path,
		Status:    status,
		Additions: adds,
		Deletions: dels,
	}
}
