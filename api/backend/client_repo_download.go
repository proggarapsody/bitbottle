package backend

import "io"

// RepoDownloadClient is implemented only by Bitbucket Cloud clients.
// It manages repository download artifacts.
type RepoDownloadClient interface {
	ListRepoDownloads(ns, slug string, limit int) ([]RepoDownload, error)
	UploadRepoDownload(ns, slug, name string, body io.Reader) (RepoDownload, error)
	DownloadRepoDownload(ns, slug, name string, out io.Writer) error
	DeleteRepoDownload(ns, slug, name string) error
}

// FeatureRepoDownloads names the repo downloads capability for typed-error reporting.
const FeatureRepoDownloads Feature = "repo_downloads"

// AsRepoDownloadClient returns the RepoDownloadClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsRepoDownloadClient(c Client, host string) (RepoDownloadClient, error) {
	return requireFeature[RepoDownloadClient](c, host, specFor(FeatureRepoDownloads))
}
