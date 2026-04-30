# Scenario: `config` and `alias`

**Backend:** Both (run once per backend — behavior is host-agnostic; running
twice catches accidental host coupling).

## Prerequisites

- Authenticated against at least one host.
- `~/.config/bitbottle/config.yml` may or may not exist; do **not** delete it
  — back it up first if you care about the contents:
  ```bash
  cp ~/.config/bitbottle/config.yml ~/.config/bitbottle/config.yml.bak 2>/dev/null || true
  ```

## Steps

### 1. `config set` writes a key

```bash
bitbottle config set git_protocol https
```

Exit code: `0`. No stdout. `~/.config/bitbottle/config.yml` now contains
`git_protocol: https`.

### 2. `config get` reads it back

```bash
bitbottle config get git_protocol
```

Stdout: `https` followed by newline. Exit code: `0`.

### 3. `config list` shows all keys

```bash
bitbottle config list
```

Stdout includes a line beginning with `git_protocol`. Exit code: `0`.

### 4. `config get` for an unknown key

```bash
bitbottle config get does-not-exist
```

Exit code: non-zero. stderr names the missing key.

### 5. `alias set` registers a shortcut

```bash
bitbottle alias set prs 'pr list --state open'
```

Exit code: `0`. stderr/stdout confirms the alias was created.

### 6. `alias list` shows it

```bash
bitbottle alias list
```

Stdout contains `prs:` (or equivalent) followed by `pr list --state open`.

### 7. The alias resolves and runs

```bash
bitbottle prs --limit 1
```

Same output as `bitbottle pr list --state open --limit 1`. Exit code: `0`.

### 8. Alias with extra args appended

```bash
bitbottle alias set viewme 'pr view'
bitbottle viewme 1 --web
```

Behaves identically to `bitbottle pr view 1 --web`.

### 9. `alias delete` removes it

```bash
bitbottle alias delete prs
bitbottle alias list
```

`prs` is no longer in the list. Exit code: `0`.

### 10. Invoking a deleted alias falls through

```bash
bitbottle prs
```

Exit code: non-zero. stderr names `prs` as an unknown command.

## Cleanup

```bash
bitbottle alias delete viewme 2>/dev/null || true
# Restore prior config if you backed it up
mv ~/.config/bitbottle/config.yml.bak ~/.config/bitbottle/config.yml 2>/dev/null || true
```
