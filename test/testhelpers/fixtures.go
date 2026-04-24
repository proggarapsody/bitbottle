package testhelpers

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
