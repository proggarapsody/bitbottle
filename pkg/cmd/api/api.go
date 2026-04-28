// Package api implements `bitbottle api` — a generic Bitbucket REST passthrough
// modeled on `gh api`. The user supplies the full Bitbucket-relative path
// (e.g. `2.0/user` for Cloud, `rest/api/1.0/projects/X/repos/Y` for Server);
// the command prepends only scheme + host. Auth and TLS settings are taken
// from the resolved host's stored config.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// options bundles the flag values for `bitbottle api`.
type options struct {
	hostname     string
	method       string
	headers      []string
	typedFields  []string // -F: JSON-typed (numbers, booleans, @file)
	stringFields []string // -f: always strings
	input        string   // --input <file|->
	jq           string   // --jq filter expression
	paginate     bool     // --paginate
}

// NewCmdAPI builds the `bitbottle api` cobra command.
func NewCmdAPI(f *factory.Factory) *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "api <endpoint>",
		Short: "Make an authenticated Bitbucket API request",
		Long: `Make an authenticated HTTP request to the Bitbucket REST API and print the response.

The endpoint is a Bitbucket-relative path. Cloud paths begin with "2.0/"
(e.g. "2.0/user"); Server / Data Center paths begin with "rest/api/1.0/"
(e.g. "rest/api/1.0/projects/PROJ/repos/REPO"). The current host's scheme,
hostname, auth token, and TLS settings are applied automatically.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPI(cmd, f, args[0], opts)
		},
	}
	cmd.Flags().StringVar(&opts.hostname, "hostname", "", "Bitbucket hostname (overrides default)")
	cmd.Flags().StringVarP(&opts.method, "method", "X", "", "HTTP method (default GET, or POST when a body is provided)")
	cmd.Flags().StringArrayVarP(&opts.headers, "header", "H", nil, "Add an HTTP request header in `key:value` format")
	cmd.Flags().StringArrayVarP(&opts.typedFields, "field", "F", nil, "Add a typed JSON body field (`key=value`); booleans, numbers, and @file are auto-detected")
	cmd.Flags().StringArrayVarP(&opts.stringFields, "raw-field", "f", nil, "Add a string-valued JSON body field (`key=value`)")
	cmd.Flags().StringVar(&opts.input, "input", "", "Read raw request body from `file` (or `-` for stdin)")
	cmd.Flags().StringVarP(&opts.jq, "jq", "q", "", "Filter the JSON response with a `jq` expression")
	cmd.Flags().BoolVar(&opts.paginate, "paginate", false, "Follow paginated responses and merge `values` arrays")
	return cmd
}

func runAPI(cmd *cobra.Command, f *factory.Factory, endpoint string, opts *options) error {
	hostnameFlag := opts.hostname
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	host := hostnameFlag
	if host == "" {
		hosts := cfg.Hosts()
		switch len(hosts) {
		case 0:
			return fmt.Errorf("not authenticated; run `bitbottle auth login` first")
		case 1:
			host = hosts[0]
		default:
			return fmt.Errorf("multiple hosts configured; use --hostname to specify one")
		}
	}
	hostCfg, ok := cfg.Get(host)
	if !ok {
		return fmt.Errorf("not logged into %s", host)
	}

	hc, err := f.HTTPClient(host)
	if err != nil {
		return err
	}

	expandedEndpoint, err := expandVars(f, endpoint)
	if err != nil {
		return err
	}

	body, contentType, err := buildBody(f, opts)
	if err != nil {
		return err
	}

	method := opts.method
	if method == "" {
		if body != nil {
			method = http.MethodPost
		} else {
			method = http.MethodGet
		}
	}

	url := "https://" + host + "/" + strings.TrimPrefix(expandedEndpoint, "/")
	req, err := http.NewRequestWithContext(cmd.Context(), method, url, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if hostCfg.OAuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+hostCfg.OAuthToken)
	}
	for _, h := range opts.headers {
		k, v, ok := strings.Cut(h, ":")
		if !ok {
			return fmt.Errorf("invalid header %q: expected key:value", h)
		}
		req.Header.Set(strings.TrimSpace(k), strings.TrimSpace(v))
	}

	if opts.paginate {
		return runPaginated(cmd, f, hc, req, opts)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if opts.jq != "" && resp.StatusCode < 400 {
		if err := applyJQ(f.IOStreams.Out, respBody, opts.jq); err != nil {
			return err
		}
	} else {
		if _, err := f.IOStreams.Out.Write(respBody); err != nil {
			return err
		}
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// runPaginated walks a Bitbucket paginated collection (Cloud `next` URLs or
// Server `nextPageStart` query params), merges every page's `values` array
// into a single JSON array, and emits it (optionally jq-filtered).
//
// Paginated body flags (-F/-f/--input) are not supported because Bitbucket
// pagination is GET-only. The caller's request method is preserved for the
// first call but Cloud's `next` URL is always GET; Server's nextPageStart
// follow-up is also GET.
func runPaginated(cmd *cobra.Command, f *factory.Factory, hc factory.HTTPClient, firstReq *http.Request, opts *options) error {
	var aggregate []any
	currentReq := firstReq

	for {
		resp, err := hc.Do(currentReq)
		if err != nil {
			return err
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close() //nolint:errcheck
		if readErr != nil {
			return readErr
		}
		if resp.StatusCode >= 400 {
			// Surface the page that failed so users can debug.
			_, _ = f.IOStreams.Out.Write(body)
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		var page paginatedPage
		if err := json.Unmarshal(body, &page); err != nil {
			return fmt.Errorf("--paginate: response is not a paginated JSON object: %w", err)
		}
		aggregate = append(aggregate, page.Values...)

		// Cloud-style: explicit next URL.
		if page.Next != "" {
			next, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, page.Next, nil)
			if err != nil {
				return err
			}
			copyAuthAndCustomHeaders(next, firstReq)
			currentReq = next
			continue
		}

		// Server-style: walk via nextPageStart while !isLastPage. We can detect
		// Server-flavored responses by the presence of the IsLastPage marker;
		// json.Unmarshal of a Cloud response leaves IsLastPage at its zero value.
		if page.hasIsLastPage && !page.IsLastPage {
			nextURL := *currentReq.URL
			q := nextURL.Query()
			q.Set("start", strconv.FormatInt(page.NextPageStart, 10))
			nextURL.RawQuery = q.Encode()

			next, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, nextURL.String(), nil)
			if err != nil {
				return err
			}
			copyAuthAndCustomHeaders(next, firstReq)
			currentReq = next
			continue
		}

		break
	}

	encoded, err := json.Marshal(aggregate)
	if err != nil {
		return err
	}
	if opts.jq != "" {
		return applyJQ(f.IOStreams.Out, encoded, opts.jq)
	}
	_, err = fmt.Fprintln(f.IOStreams.Out, string(encoded))
	return err
}

// paginatedPage models the relevant subset of both Cloud and Server collection
// responses. Unknown fields are ignored. hasIsLastPage is set via custom
// unmarshaling so we can distinguish "Server response with isLastPage:true"
// from "Cloud response that omits the field".
type paginatedPage struct {
	Values        []any `json:"values"`
	Next          string
	IsLastPage    bool
	NextPageStart int64
	hasIsLastPage bool
}

func (p *paginatedPage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Values        []any           `json:"values"`
		Next          string          `json:"next"`
		IsLastPage    *bool           `json:"isLastPage"`
		NextPageStart json.Number     `json:"nextPageStart"`
		_             json.RawMessage // silence unkeyed-fields lint
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.Values = raw.Values
	p.Next = raw.Next
	if raw.IsLastPage != nil {
		p.hasIsLastPage = true
		p.IsLastPage = *raw.IsLastPage
	}
	if raw.NextPageStart != "" {
		n, err := raw.NextPageStart.Int64()
		if err != nil {
			return fmt.Errorf("nextPageStart: %w", err)
		}
		p.NextPageStart = n
	}
	return nil
}

func copyAuthAndCustomHeaders(dst, src *http.Request) {
	if v := src.Header.Get("Authorization"); v != "" {
		dst.Header.Set("Authorization", v)
	}
	for k, vs := range src.Header {
		if k == "Authorization" {
			continue
		}
		for _, v := range vs {
			dst.Header.Add(k, v)
		}
	}
}

// repoVarNames are the placeholder tokens replaced from the resolved base repo.
// project/slug match RepoRef field names (used by Server / DC paths);
// workspace/repo_slug are Cloud-flavored aliases for the same values, since
// Bitbucket Cloud calls them "workspace" and "repo_slug" in its docs.
var repoVarNames = []string{"{project}", "{slug}", "{workspace}", "{repo_slug}"}

// expandVars substitutes {workspace}/{repo_slug}/{project}/{slug} in endpoint
// with values from the resolved base repo. Vars are looked up lazily so paths
// without any placeholders work outside a git checkout.
func expandVars(f *factory.Factory, endpoint string) (string, error) {
	needsRepo := false
	for _, v := range repoVarNames {
		if strings.Contains(endpoint, v) {
			needsRepo = true
			break
		}
	}
	if !needsRepo {
		return endpoint, nil
	}

	if f.BaseRepo == nil {
		return "", fmt.Errorf("path contains repo placeholder but no base repo is configured")
	}
	ref, err := f.BaseRepo()
	if err != nil {
		// Surface placeholder name to make the failure mode obvious.
		token := firstPresent(endpoint, repoVarNames)
		return "", fmt.Errorf("cannot resolve %s: %w", token, err)
	}

	r := strings.NewReplacer(
		"{project}", ref.Project,
		"{workspace}", ref.Project,
		"{slug}", ref.Slug,
		"{repo_slug}", ref.Slug,
	)
	return r.Replace(endpoint), nil
}

func firstPresent(s string, tokens []string) string {
	for _, t := range tokens {
		if strings.Contains(s, t) {
			return t
		}
	}
	return ""
}

// buildBody assembles the request body from -F/-f/--input flags. Returns
// (nil, "", nil) when no body flags are set. -F and -f populate a JSON object;
// --input is mutually exclusive with field flags and streams the file's bytes
// verbatim (with no Content-Type set unless the user supplies -H).
func buildBody(f *factory.Factory, opts *options) (io.Reader, string, error) {
	if opts.input != "" {
		if len(opts.typedFields) > 0 || len(opts.stringFields) > 0 {
			return nil, "", fmt.Errorf("--input is mutually exclusive with --field/--raw-field")
		}
		var r io.Reader
		if opts.input == "-" {
			r = f.IOStreams.In
		} else {
			data, err := os.ReadFile(opts.input)
			if err != nil {
				return nil, "", fmt.Errorf("read --input: %w", err)
			}
			r = bytes.NewReader(data)
		}
		return r, "", nil
	}

	if len(opts.typedFields) == 0 && len(opts.stringFields) == 0 {
		return nil, "", nil
	}

	payload := make(map[string]any, len(opts.typedFields)+len(opts.stringFields))
	for _, raw := range opts.typedFields {
		k, v, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, "", fmt.Errorf("invalid -F %q: expected key=value", raw)
		}
		val, err := parseTypedValue(v)
		if err != nil {
			return nil, "", err
		}
		payload[k] = val
	}
	for _, raw := range opts.stringFields {
		k, v, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, "", fmt.Errorf("invalid -f %q: expected key=value", raw)
		}
		payload[k] = v
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(encoded), "application/json", nil
}

// applyJQ runs the user-supplied jq expression against the JSON response body
// and writes one result per line to w. Each result is JSON-encoded so that
// scalars (strings, numbers, booleans, null) round-trip cleanly into pipes.
func applyJQ(w io.Writer, respBody []byte, expr string) error {
	var input any
	if err := json.Unmarshal(respBody, &input); err != nil {
		return fmt.Errorf("--jq: response is not valid JSON: %w", err)
	}
	q, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("--jq: %w", err)
	}
	iter := q.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if e, ok := v.(error); ok {
			return fmt.Errorf("--jq: %w", e)
		}
		out, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return err
		}
	}
	return nil
}

// parseTypedValue converts a -F right-hand side into its JSON-typed Go value.
// Recognised forms: "true"/"false" → bool, integer → int64, "@filename" →
// file contents as string, "@-" → stdin contents as string. Anything else is
// kept as a string.
func parseTypedValue(v string) (any, error) {
	switch v {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	if n, err := strconv.ParseInt(v, 10, 64); err == nil {
		return n, nil
	}
	if strings.HasPrefix(v, "@") {
		path := strings.TrimPrefix(v, "@")
		if path == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("read stdin for -F: %w", err)
			}
			return string(data), nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s for -F: %w", path, err)
		}
		return string(data), nil
	}
	return v, nil
}
