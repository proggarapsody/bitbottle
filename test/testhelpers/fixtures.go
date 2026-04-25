package testhelpers

import "github.com/proggarapsody/bitbottle/api/backend"

// RepoOption mutates a Bitbucket repository fixture.
type RepoOption func(map[string]any)

// RepoFactory returns a default Bitbucket repository object with optional
// functional overrides applied in order.
func RepoFactory(opts ...RepoOption) map[string]any {
	repo := map[string]any{
		"id":   1,
		"slug": "default-repo",
		"name": "default-repo",
		"project": map[string]any{
			"key":  "PROJ",
			"name": "Project",
		},
		"scmId":  "git",
		"state":  "AVAILABLE",
		"public": false,
		"links": map[string]any{
			"clone": []map[string]any{
				{"href": "https://bitbucket.example.com/scm/proj/default-repo.git", "name": "http"},
				{"href": "ssh://git@bitbucket.example.com:7999/proj/default-repo.git", "name": "ssh"},
			},
			"self": []map[string]any{
				{"href": "https://bitbucket.example.com/projects/PROJ/repos/default-repo/browse"},
			},
		},
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

// RepoWithSlug sets the slug and name fields of the repo.
func RepoWithSlug(slug string) RepoOption {
	return func(r map[string]any) {
		r["slug"] = slug
		r["name"] = slug
	}
}

// RepoWithProject sets the nested project.key field.
func RepoWithProject(key string) RepoOption {
	return func(r map[string]any) {
		proj, _ := r["project"].(map[string]any)
		if proj == nil {
			proj = map[string]any{}
		}
		proj["key"] = key
		r["project"] = proj
	}
}

// RepoWithPublic sets the public flag.
func RepoWithPublic(pub bool) RepoOption {
	return func(r map[string]any) {
		r["public"] = pub
	}
}

// PROption mutates a Bitbucket pull request fixture.
type PROption func(map[string]any)

// PRFactory returns a default Bitbucket pull request object with optional
// functional overrides applied in order.
func PRFactory(opts ...PROption) map[string]any {
	pr := map[string]any{
		"id":          1,
		"title":       "Default PR title",
		"description": "",
		"state":       "OPEN",
		"author": map[string]any{
			"user": map[string]any{
				"slug":        "alice",
				"displayName": "Alice",
			},
			"role": "AUTHOR",
		},
		"fromRef": map[string]any{
			"id":        "refs/heads/feat/my-feature",
			"displayId": "feat/my-feature",
		},
		"toRef": map[string]any{
			"id":        "refs/heads/main",
			"displayId": "main",
		},
		"draft": false,
	}
	for _, opt := range opts {
		opt(pr)
	}
	return pr
}

// PRWithID sets the PR id.
func PRWithID(id int) PROption {
	return func(p map[string]any) { p["id"] = id }
}

// PRWithTitle sets the PR title.
func PRWithTitle(title string) PROption {
	return func(p map[string]any) { p["title"] = title }
}

// PRWithState sets the PR state (OPEN, MERGED, DECLINED).
func PRWithState(state string) PROption {
	return func(p map[string]any) { p["state"] = state }
}

// PRWithDraft sets the draft flag.
func PRWithDraft(draft bool) PROption {
	return func(p map[string]any) { p["draft"] = draft }
}

// PRWithBranch sets fromRef to the given source branch.
func PRWithBranch(branch string) PROption {
	return func(p map[string]any) {
		p["fromRef"] = map[string]any{
			"id":        "refs/heads/" + branch,
			"displayId": branch,
		}
	}
}

// PRWithAuthor sets author.user.slug.
func PRWithAuthor(slug string) PROption {
	return func(p map[string]any) {
		author, _ := p["author"].(map[string]any)
		if author == nil {
			author = map[string]any{"role": "AUTHOR"}
		}
		user, _ := author["user"].(map[string]any)
		if user == nil {
			user = map[string]any{}
		}
		user["slug"] = slug
		author["user"] = user
		p["author"] = author
	}
}

// PRWithBaseBranch sets toRef to the given destination branch.
func PRWithBaseBranch(branch string) PROption {
	return func(p map[string]any) {
		p["toRef"] = map[string]any{
			"id":        "refs/heads/" + branch,
			"displayId": branch,
		}
	}
}

// --- Cloud fixture factories ---

// CloudRepoOption mutates a Bitbucket Cloud repository fixture.
type CloudRepoOption func(map[string]any)

// CloudRepoFactory returns a default Bitbucket Cloud repository object.
func CloudRepoFactory(opts ...CloudRepoOption) map[string]any {
	repo := map[string]any{
		"type":      "repository",
		"full_name": "myworkspace/default-repo",
		"slug":      "default-repo",
		"name":      "default-repo",
		"scm":       "git",
		"is_private": true,
		"links": map[string]any{
			"html": map[string]any{
				"href": "https://bitbucket.org/myworkspace/default-repo",
			},
			"clone": []map[string]any{},
		},
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

// CloudRepoWithSlug sets the slug, name, and full_name of the cloud repo.
func CloudRepoWithSlug(workspace, slug string) CloudRepoOption {
	return func(r map[string]any) {
		r["slug"] = slug
		r["name"] = slug
		r["full_name"] = workspace + "/" + slug
	}
}

// CloudPROption mutates a Bitbucket Cloud pull request fixture.
type CloudPROption func(map[string]any)

// CloudPRFactory returns a default Bitbucket Cloud pull request object.
func CloudPRFactory(opts ...CloudPROption) map[string]any {
	pr := map[string]any{
		"type":        "pullrequest",
		"id":          1,
		"title":       "Default Cloud PR title",
		"description": "",
		"state":       "OPEN",
		"draft":       false,
		"author": map[string]any{
			"display_name": "Alice",
			"account_id":   "alice-uuid",
			"nickname":     "alice",
		},
		"source": map[string]any{
			"branch": map[string]any{"name": "feat/my-feature"},
		},
		"destination": map[string]any{
			"branch": map[string]any{"name": "main"},
		},
		"links": map[string]any{
			"html": map[string]any{
				"href": "https://bitbucket.org/myworkspace/default-repo/pull-requests/1",
			},
		},
	}
	for _, opt := range opts {
		opt(pr)
	}
	return pr
}

// CloudPRWithID sets the Cloud PR id.
func CloudPRWithID(id int) CloudPROption {
	return func(p map[string]any) { p["id"] = id }
}

// CloudPRWithTitle sets the Cloud PR title.
func CloudPRWithTitle(title string) CloudPROption {
	return func(p map[string]any) { p["title"] = title }
}

// CloudPRWithState sets the Cloud PR state.
func CloudPRWithState(state string) CloudPROption {
	return func(p map[string]any) { p["state"] = state }
}

// CloudPRWithDraft sets the draft flag on a Cloud PR.
func CloudPRWithDraft(draft bool) CloudPROption {
	return func(p map[string]any) { p["draft"] = draft }
}

// CloudPRWithAuthor sets the Cloud PR author account_id and display_name.
func CloudPRWithAuthor(accountID, displayName string) CloudPROption {
	return func(p map[string]any) {
		p["author"] = map[string]any{
			"account_id":   accountID,
			"display_name": displayName,
			"nickname":     accountID,
		}
	}
}

// --- backend domain factories ---

// BackendRepoOption mutates a backend.Repository.
type BackendRepoOption func(*backend.Repository)

// BackendRepoFactory returns a default backend.Repository with options applied.
func BackendRepoFactory(opts ...BackendRepoOption) backend.Repository {
	r := backend.Repository{
		Slug:      "default-repo",
		Name:      "default-repo",
		Namespace: "PROJ",
		SCM:       "git",
		WebURL:    "https://bitbucket.example.com/projects/PROJ/repos/default-repo/browse",
	}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

// BackendPROption mutates a backend.PullRequest.
type BackendPROption func(*backend.PullRequest)

// BackendPRFactory returns a default backend.PullRequest with options applied.
func BackendPRFactory(opts ...BackendPROption) backend.PullRequest {
	pr := backend.PullRequest{
		ID:         1,
		Title:      "Default PR title",
		State:      "OPEN",
		Draft:      false,
		FromBranch: "feat/my-feature",
		ToBranch:   "main",
		Author: backend.User{
			Slug:        "alice",
			DisplayName: "Alice",
		},
	}
	for _, opt := range opts {
		opt(&pr)
	}
	return pr
}

// BackendPRWithID sets the ID on a backend.PullRequest.
func BackendPRWithID(id int) BackendPROption {
	return func(p *backend.PullRequest) { p.ID = id }
}

// BackendPRWithWebURL sets the WebURL on a backend.PullRequest.
func BackendPRWithWebURL(url string) BackendPROption {
	return func(p *backend.PullRequest) { p.WebURL = url }
}

// BackendPRWithFromBranch sets FromBranch on a backend.PullRequest.
func BackendPRWithFromBranch(branch string) BackendPROption {
	return func(p *backend.PullRequest) { p.FromBranch = branch }
}

// BackendPRWithTitle sets Title on a backend.PullRequest.
func BackendPRWithTitle(title string) BackendPROption {
	return func(p *backend.PullRequest) { p.Title = title }
}

// BackendPRWithState sets State on a backend.PullRequest.
func BackendPRWithState(state string) BackendPROption {
	return func(p *backend.PullRequest) { p.State = state }
}

// BackendRepoWithWebURL sets the WebURL on a backend.Repository.
func BackendRepoWithWebURL(url string) BackendRepoOption {
	return func(r *backend.Repository) { r.WebURL = url }
}

// BackendRepoWithSlug sets the slug on a backend.Repository.
func BackendRepoWithSlug(slug string) BackendRepoOption {
	return func(r *backend.Repository) {
		r.Slug = slug
		r.Name = slug
	}
}

// BackendUserOption mutates a backend.User.
type BackendUserOption func(*backend.User)

// BackendUserFactory returns a default backend.User with options applied.
func BackendUserFactory(opts ...BackendUserOption) backend.User {
	u := backend.User{
		Slug:        "alice",
		DisplayName: "Alice Example",
	}
	for _, opt := range opts {
		opt(&u)
	}
	return u
}

// BackendUserWithSlug sets the slug on a backend.User.
func BackendUserWithSlug(slug string) BackendUserOption {
	return func(u *backend.User) { u.Slug = slug }
}
