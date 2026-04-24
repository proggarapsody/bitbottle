package cloud_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

// buildCloudReposBody builds a JSON response body containing n cloud
// repositories in the paged envelope used by Bitbucket Cloud.
func buildCloudReposBody(n int) []byte {
	type link struct {
		Href string `json:"href"`
	}
	type links struct {
		HTML link `json:"html"`
	}
	type repo struct {
		Type     string `json:"type"`
		FullName string `json:"full_name"`
		Slug     string `json:"slug"`
		Name     string `json:"name"`
		SCM      string `json:"scm"`
		Links    links  `json:"links"`
	}
	type envelope struct {
		Pagelen int    `json:"pagelen"`
		Page    int    `json:"page"`
		Size    int    `json:"size"`
		Values  []repo `json:"values"`
	}
	vals := make([]repo, n)
	for i := 0; i < n; i++ {
		slug := fmt.Sprintf("repo-%d", i)
		vals[i] = repo{
			Type:     "repository",
			FullName: "myworkspace/" + slug,
			Slug:     slug,
			Name:     slug,
			SCM:      "git",
			Links: links{
				HTML: link{Href: "https://bitbucket.org/myworkspace/" + slug},
			},
		}
	}
	b, _ := json.Marshal(envelope{Pagelen: n, Page: 1, Size: n, Values: vals})
	return b
}

// buildCloudPRsBody builds a JSON response body containing n cloud PRs in the
// paged envelope used by Bitbucket Cloud.
func buildCloudPRsBody(n int) []byte {
	type branch struct {
		Name string `json:"name"`
	}
	type ref struct {
		Branch branch `json:"branch"`
	}
	type author struct {
		DisplayName string `json:"display_name"`
		AccountID   string `json:"account_id"`
	}
	type link struct {
		Href string `json:"href"`
	}
	type links struct {
		HTML link `json:"html"`
	}
	type pr struct {
		Type        string `json:"type"`
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		State       string `json:"state"`
		Draft       bool   `json:"draft"`
		Author      author `json:"author"`
		Source      ref    `json:"source"`
		Destination ref    `json:"destination"`
		Links       links  `json:"links"`
	}
	type envelope struct {
		Pagelen int  `json:"pagelen"`
		Page    int  `json:"page"`
		Size    int  `json:"size"`
		Values  []pr `json:"values"`
	}
	vals := make([]pr, n)
	for i := 0; i < n; i++ {
		vals[i] = pr{
			Type:        "pullrequest",
			ID:          i + 1,
			Title:       fmt.Sprintf("PR %d", i+1),
			Description: "description",
			State:       "OPEN",
			Draft:       false,
			Author: author{
				DisplayName: "Alice",
				AccountID:   "alice-uuid",
			},
			Source:      ref{Branch: branch{Name: fmt.Sprintf("feat/x-%d", i)}},
			Destination: ref{Branch: branch{Name: "main"}},
			Links: links{
				HTML: link{Href: fmt.Sprintf("https://bitbucket.org/myworkspace/my-service/pull-requests/%d", i+1)},
			},
		}
	}
	b, _ := json.Marshal(envelope{Pagelen: n, Page: 1, Size: n, Values: vals})
	return b
}

func BenchmarkCloudClient_ListRepos_Decode100(b *testing.B) {
	body := buildCloudReposBody(100)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		repos, err := client.ListRepos(100)
		if err != nil {
			b.Fatal(err)
		}
		if len(repos) != 100 {
			b.Fatalf("expected 100 repos, got %d", len(repos))
		}
	}
}

func BenchmarkCloudClient_ListPRs_Decode100(b *testing.B) {
	body := buildCloudPRsBody(100)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		prs, err := client.ListPRs("myworkspace", "my-service", "OPEN", 100)
		if err != nil {
			b.Fatal(err)
		}
		if len(prs) != 100 {
			b.Fatalf("expected 100 PRs, got %d", len(prs))
		}
	}
}
