package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/proggarapsody/bitbottle/api/server"
)

// buildServerReposBody builds a JSON response body containing n server repos
// in the paged envelope used by Bitbucket Data Center.
func buildServerReposBody(n int) []byte {
	type selfLink struct {
		Href string `json:"href"`
	}
	type links struct {
		Self []selfLink `json:"self"`
	}
	type project struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}
	type repo struct {
		ID      int     `json:"id"`
		Slug    string  `json:"slug"`
		Name    string  `json:"name"`
		Project project `json:"project"`
		ScmID   string  `json:"scmId"`
		State   string  `json:"state"`
		Links   links   `json:"links"`
	}
	type envelope struct {
		Values     []repo `json:"values"`
		Size       int    `json:"size"`
		IsLastPage bool   `json:"isLastPage"`
		Start      int    `json:"start"`
		Limit      int    `json:"limit"`
	}
	vals := make([]repo, n)
	for i := 0; i < n; i++ {
		slug := fmt.Sprintf("repo-%d", i)
		vals[i] = repo{
			ID:      i + 1,
			Slug:    slug,
			Name:    slug,
			Project: project{Key: "MYPROJ", Name: "My Project"},
			ScmID:   "git",
			State:   "AVAILABLE",
			Links: links{Self: []selfLink{
				{Href: "https://bb.example.com/projects/MYPROJ/repos/" + slug + "/browse"},
			}},
		}
	}
	b, _ := json.Marshal(envelope{Values: vals, Size: n, IsLastPage: true, Start: 0, Limit: n})
	return b
}

// buildServerPRsBody builds a JSON response body containing n server PRs in
// the paged envelope used by Bitbucket Data Center.
func buildServerPRsBody(n int) []byte {
	type selfLink struct {
		Href string `json:"href"`
	}
	type links struct {
		Self []selfLink `json:"self"`
	}
	type user struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"displayName"`
	}
	type author struct {
		User user   `json:"user"`
		Role string `json:"role"`
	}
	type ref struct {
		ID        string `json:"id"`
		DisplayID string `json:"displayId"`
	}
	type pr struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		State       string `json:"state"`
		Draft       bool   `json:"draft"`
		Author      author `json:"author"`
		FromRef     ref    `json:"fromRef"`
		ToRef       ref    `json:"toRef"`
		Links       links  `json:"links"`
	}
	type envelope struct {
		Values     []pr `json:"values"`
		Size       int  `json:"size"`
		IsLastPage bool `json:"isLastPage"`
		Start      int  `json:"start"`
		Limit      int  `json:"limit"`
	}
	vals := make([]pr, n)
	for i := 0; i < n; i++ {
		vals[i] = pr{
			ID:          i + 1,
			Title:       fmt.Sprintf("PR %d", i+1),
			Description: "description",
			State:       "OPEN",
			Draft:       false,
			Author: author{
				User: user{Slug: "alice", DisplayName: "Alice"},
				Role: "AUTHOR",
			},
			FromRef: ref{ID: fmt.Sprintf("refs/heads/feat/x-%d", i), DisplayID: fmt.Sprintf("feat/x-%d", i)},
			ToRef:   ref{ID: "refs/heads/main", DisplayID: "main"},
			Links: links{Self: []selfLink{
				{Href: fmt.Sprintf("https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/%d", i+1)},
			}},
		}
	}
	b, _ := json.Marshal(envelope{Values: vals, Size: n, IsLastPage: true, Start: 0, Limit: n})
	return b
}

func BenchmarkServerClient_ListRepos_Decode100(b *testing.B) {
	body := buildServerReposBody(100)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

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

func BenchmarkServerClient_ListPRs_Decode100(b *testing.B) {
	body := buildServerPRsBody(100)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		prs, err := client.ListPRs("MYPROJ", "my-service", "OPEN", 100)
		if err != nil {
			b.Fatal(err)
		}
		if len(prs) != 100 {
			b.Fatalf("expected 100 PRs, got %d", len(prs))
		}
	}
}
