package bbinstance

import (
	"fmt"
	"strconv"
	"strings"
)

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
