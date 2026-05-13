# bitbottle perms — Permissions Management (Server / DC only)

> Manage user and group permissions for projects and repositories.
> Requires `PROJECT_ADMIN` on the target project. Returns
> `host.unsupported` error when called against Bitbucket Cloud.

---

## Command reference

### Project permissions

```bash
# List all grants (users + groups merged, sorted ADMIN → WRITE → READ)
bitbottle --hostname HOST perms project list PROJECT
bitbottle --hostname HOST perms project list PROJECT --json permission,subject
bitbottle --hostname HOST perms project list PROJECT --jq '.[] | select(.permission == "PROJECT_ADMIN")'

# Grant a user/group a project permission
bitbottle --hostname HOST perms project grant PROJECT PERM --user SLUG
bitbottle --hostname HOST perms project grant PROJECT PERM --group "group name"
# Downgrade (ADMIN→READ) warns on TTY; --force skips the prompt
bitbottle --hostname HOST perms project grant PROJECT PROJECT_READ --user alice --force

# Revoke a user/group from a project
bitbottle --hostname HOST perms project revoke PROJECT --user SLUG
bitbottle --hostname HOST perms project revoke PROJECT --group "group name"
```

### Repo permissions

```bash
# Arg is PROJECT/REPO (slash-separated)
bitbottle --hostname HOST perms repo list PROJECT/REPO
bitbottle --hostname HOST perms repo list PROJECT/REPO --json

bitbottle --hostname HOST perms repo grant PROJECT/REPO PERM --user SLUG
bitbottle --hostname HOST perms repo grant PROJECT/REPO PERM --group "group name"

bitbottle --hostname HOST perms repo revoke PROJECT/REPO --user SLUG
bitbottle --hostname HOST perms repo revoke PROJECT/REPO --group "group name"
```

---

## Valid permission levels

| Scope   | Values |
|---------|--------|
| Project | `PROJECT_READ`, `PROJECT_WRITE`, `PROJECT_ADMIN` |
| Repo    | `REPO_READ`, `REPO_WRITE`, `REPO_ADMIN` |

---

## MCP tools

All 6 tools are Server/DC-only. Pass `hostname` when multiple hosts are configured.

| Tool | Required params | Optional |
|------|-----------------|---------|
| `list_project_permissions` | `project` | `hostname` |
| `grant_project_permission` | `project`, `permission`, one of `user`/`group` | `hostname` |
| `revoke_project_permission` | `project`, one of `user`/`group` | `hostname` |
| `list_repo_permissions` | `project`, `slug` | `hostname` |
| `grant_repo_permission` | `project`, `slug`, `permission`, one of `user`/`group` | `hostname` |
| `revoke_repo_permission` | `project`, `slug`, one of `user`/`group` | `hostname` |

---

## Backend interface

```go
// api/backend/client_permissions.go
type PermissionsClient interface {
    ListProjectPermissions(ctx context.Context, project string) ([]PermissionGrant, error)
    GrantProjectPermission(ctx context.Context, project string, subject PermissionSubject, perm string) error
    RevokeProjectPermission(ctx context.Context, project string, subject PermissionSubject) error
    ListRepoPermissions(ctx context.Context, project, slug string) ([]PermissionGrant, error)
    GrantRepoPermission(ctx context.Context, project, slug string, subject PermissionSubject, perm string) error
    RevokeRepoPermission(ctx context.Context, project, slug string, subject PermissionSubject) error
}

// Optional-interface accessor (returns ErrUnsupportedOnHost on Cloud)
pc, err := backend.AsPermissionsClient(client, hostname)
```

```go
// api/backend/types.go
type PermissionSubject struct {
    Kind        string // "user" | "group"
    Slug        string // user slug (Kind=user)
    Name        string // group name (Kind=group)
    DisplayName string // populated on read; ignored on write
}
type PermissionGrant struct {
    Subject    PermissionSubject
    Permission string
}
```

---

## Server API endpoints

| Action | Method | Path |
|--------|--------|------|
| List project users | GET | `/rest/api/1.0/projects/{key}/permissions/users` |
| List project groups | GET | `/rest/api/1.0/projects/{key}/permissions/groups` |
| Grant project user | PUT | `/rest/api/1.0/projects/{key}/permissions/users?name={slug}&permission={perm}` |
| Grant project group | PUT | `/rest/api/1.0/projects/{key}/permissions/groups?name={name}&permission={perm}` |
| Revoke project user | DELETE | `/rest/api/1.0/projects/{key}/permissions/users?name={slug}` |
| Revoke project group | DELETE | `/rest/api/1.0/projects/{key}/permissions/groups?name={name}` |
| List repo users | GET | `/rest/api/1.0/projects/{key}/repos/{slug}/permissions/users` |
| List repo groups | GET | `/rest/api/1.0/projects/{key}/repos/{slug}/permissions/groups` |
| Grant repo user | PUT | `/rest/api/1.0/projects/{key}/repos/{slug}/permissions/users?name=…` |
| Grant repo group | PUT | `/rest/api/1.0/projects/{key}/repos/{slug}/permissions/groups?name=…` |
| Revoke repo user | DELETE | `/rest/api/1.0/projects/{key}/repos/{slug}/permissions/users?name=…` |
| Revoke repo group | DELETE | `/rest/api/1.0/projects/{key}/repos/{slug}/permissions/groups?name=…` |

`List*` endpoints are paginated (`isLastPage` / `nextPageStart`). Both user
and group endpoints are fetched in parallel then merged by `mergeAndSort`.

---

## Error codes

| Situation | Error |
|-----------|-------|
| Called on Cloud host | `host.unsupported` (`backend.ErrUnsupportedOnHost`) |
| Not PROJECT_ADMIN | `perms.admin_required` (403 → `backend.ErrPermission`) |
