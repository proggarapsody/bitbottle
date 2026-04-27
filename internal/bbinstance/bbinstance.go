package bbinstance

import (
	"fmt"
	"strconv"
	"strings"
)

// Recognised values for HostConfig.BackendType. Empty string means "infer
// from hostname".
const (
	BackendTypeCloud  = "cloud"
	BackendTypeServer = "server"
)

// cloudHostname is the canonical Bitbucket Cloud hostname.
const cloudHostname = "bitbucket.org"

// IsCloud returns true when the hostname or backendType indicates Bitbucket Cloud.
// Rules (in order):
//  1. backendType == "cloud"  → always true
//  2. backendType is any non-empty, non-"cloud" value (e.g. "server", "datacenter") → false
//  3. No backendType: hostname == "bitbucket.org" (exact, no port) → true
//  4. Everything else → false
func IsCloud(hostname, backendType string) bool {
	switch backendType {
	case BackendTypeCloud:
		return true
	case "":
		return hostname == cloudHostname
	default:
		return false
	}
}

// cloudAPIHostname is the canonical Bitbucket Cloud REST API hostname.
const cloudAPIHostname = "api.bitbucket.org"

// CloudRESTBase returns the Bitbucket Cloud REST API v2.0 base URL.
func CloudRESTBase() string {
	return "https://" + cloudAPIHostname + "/2.0"
}

// CloudSSHURL builds git@bitbucket.org:NAMESPACE/SLUG.git
func CloudSSHURL(namespace, slug string) string {
	return fmt.Sprintf("git@%s:%s/%s.git", cloudHostname, namespace, slug)
}

// CloudHTTPSURL builds https://bitbucket.org/NAMESPACE/SLUG.git
func CloudHTTPSURL(namespace, slug string) string {
	return fmt.Sprintf("https://%s/%s/%s.git", cloudHostname, namespace, slug)
}

// CloudWebRepoURL builds https://bitbucket.org/NAMESPACE/SLUG
func CloudWebRepoURL(namespace, slug string) string {
	return fmt.Sprintf("https://%s/%s/%s", cloudHostname, namespace, slug)
}

// CloudWebPRURL builds https://bitbucket.org/NAMESPACE/SLUG/pull-requests/ID
func CloudWebPRURL(namespace, slug string, id int) string {
	return fmt.Sprintf("https://%s/%s/%s/pull-requests/%d", cloudHostname, namespace, slug, id)
}

// SSHURL builds git@HOST:PROJECT/REPO.git
func SSHURL(host, project, slug string) string {
	return fmt.Sprintf("git@%s:%s/%s.git", host, project, slug)
}

// HTTPSURL builds https://HOST/scm/PROJECT/REPO.git
func HTTPSURL(host, project, slug string) string {
	return fmt.Sprintf("https://%s/scm/%s/%s.git", host, project, slug)
}

// WebRepoURL builds https://HOST/projects/PROJECT/repos/SLUG/browse
func WebRepoURL(host, project, slug string) string {
	return fmt.Sprintf("https://%s/projects/%s/repos/%s/browse", host, project, slug)
}

// WebPRURL builds https://HOST/projects/PROJECT/repos/SLUG/pull-requests/ID
func WebPRURL(host, project, slug string, id int) string {
	return fmt.Sprintf("https://%s/projects/%s/repos/%s/pull-requests/%d", host, project, slug, id)
}

// RESTBase builds https://HOST/rest/api/1.0
func RESTBase(host string) string {
	return fmt.Sprintf("https://%s/rest/api/1.0", host)
}

// PATManageURL returns the URL for managing Personal Access Tokens on a
// Bitbucket Server / Data Center instance.
func PATManageURL(hostname string) string {
	return fmt.Sprintf("https://%s/plugins/servlet/access-tokens/manage", hostname)
}

// CloudAppPasswordsURL returns the Bitbucket Cloud App Passwords creation URL.
func CloudAppPasswordsURL() string {
	return "https://bitbucket.org/account/settings/app-passwords/new"
}

// SupportsDraftPR returns true if version >= "7.17.0".
func SupportsDraftPR(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	if major > 7 {
		return true
	}
	return major == 7 && minor >= 17
}
