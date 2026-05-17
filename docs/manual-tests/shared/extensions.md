# Scenario: Extensions (install / list / exec / upgrade / remove)

**Backend:** N/A (local install + GitHub-hosted plugins). The extension
mechanism is host-agnostic, but `exec` may itself call out to a backend
depending on the extension.

Covers **EXT-CORE** (install + list), **EXT-RUNTIME** (exec, SHA
verification, env sanitisation), and **EXT-MGMT** (upgrade + remove).
Exercises the real plugin install path: a GitHub release tarball is
downloaded, the SHA is recorded in the lockfile, and the binary is
materialised under `~/.config/bitbottle/extensions/`.

## Prerequisites

- Network access to `github.com`.
- A working, harmless extension to install. Recommended:
  `proggarapsody/bitbottle-ext-hello` if it exists, otherwise pick any
  small public extension from the bitbottle org or another trusted source
  — record which one was used.
- A local checkout of any small Go binary you can rename to look like an
  extension (for the `--local` install path in step 7). Example:
  `make build && cp dist/bitbottle /tmp/bitbottle-hello/bin/bitbottle-hello`
  — verify your local extension's filename matches its directory name
  per the extension contract.

## Setup

```bash
# Confirm starting from a clean extension state:
bitbottle extension list
```

If extensions already exist from prior runs, leave them; only the QA
ones below are exercised.

## Steps — Remote install (EXT-CORE)

### 1. `extension install USER/REPO`

```bash
bitbottle extension install proggarapsody/bitbottle-ext-hello
```

Exit code: `0`. stderr reports the download URL, the discovered release
asset, and the SHA recorded in the lockfile. The binary lands under
`~/.config/bitbottle/extensions/hello/`.

If the example extension doesn't exist, substitute any trusted public
extension and record which one — the test is the install plumbing.

### 2. `extension list` includes the new extension

```bash
bitbottle extension list
```

Stdout is a table with columns `NAME`, `KIND` (`remote`), `VERSION`,
`PATH`. The new row is present. Exit code: `0`.

### 3. `extension exec` runs it

```bash
bitbottle extension exec hello --version 2>&1 | head -3
```

Stdout is the extension's own output (a version line, a help banner,
etc.). Exit code: depends on the extension; record what was observed.

The shorter root dispatch (`bitbottle hello --version`) should produce
the same output:

```bash
bitbottle hello --version 2>&1 | head -3
```

### 4. SHA tamper detection

Open the lockfile and corrupt the SHA for the installed extension:

```bash
cp ~/.config/bitbottle/extensions/lockfile.yml /tmp/bb-lock-bak
# Replace the recorded SHA with a clearly-wrong value:
sed -i.bak 's/sha256: [a-f0-9]*/sha256: 0000000000000000000000000000000000000000000000000000000000000000/' \
  ~/.config/bitbottle/extensions/lockfile.yml
```

Now try to exec:

```bash
bitbottle extension exec hello --version
```

Exit code: non-zero. stderr mentions a SHA mismatch / integrity failure
and refuses to run the binary. Restore:

```bash
cp /tmp/bb-lock-bak ~/.config/bitbottle/extensions/lockfile.yml
rm -f /tmp/bb-lock-bak ~/.config/bitbottle/extensions/lockfile.yml.bak
bitbottle extension exec hello --version >/dev/null
```

Last command exits `0` again.

### 5. Env sanitisation: `BB_` vars not leaked

```bash
BB_INTERNAL_DEBUG=1 BB_TEST_CLOUD_TOKEN=visible-to-bitbottle bitbottle \
  extension exec hello env 2>&1 | grep '^BB_' || echo "sanitised"
```

If the extension echoes its environment, no `BB_*` variables that are
considered secret (token vars) appear in its env. The final line prints
`sanitised` or only `BB_*` allowlisted vars (e.g. `BB_PLUGIN_*` injected
on purpose). Record the actual policy.

## Steps — Local install (EXT-CORE `--local`)

### 6. `extension install --local PATH`

Build a tiny stub extension. Create the directory layout:

```bash
mkdir -p /tmp/bitbottle-ext-noop/bin
cat > /tmp/bitbottle-ext-noop/bin/bitbottle-ext-noop <<'EOF'
#!/usr/bin/env bash
echo "noop $@"
EOF
chmod +x /tmp/bitbottle-ext-noop/bin/bitbottle-ext-noop
```

Install via symlink:

```bash
bitbottle extension install --local /tmp/bitbottle-ext-noop
```

Exit code: `0`. `extension list` now shows two rows; the new one has
`KIND=local` (or similar).

### 7. The local extension is exec-able

```bash
bitbottle extension exec ext-noop hi
```

Stdout is `noop hi`.

## Steps — Upgrade & remove (EXT-MGMT)

### 8. `extension upgrade NAME` (single)

```bash
bitbottle extension upgrade hello
```

Exit code: `0`. If already at the latest release, stderr says so. Otherwise
the new version is downloaded, the lockfile updated, and the old binary
replaced.

### 9. `extension upgrade --all`

```bash
bitbottle extension upgrade --all
```

Exit code: `0`. Each remote extension is checked. Local extensions are
skipped with a message.

### 10. `extension upgrade --force` reinstalls

```bash
bitbottle extension upgrade hello --force
```

Exit code: `0`. stderr explicitly notes the reinstall path.

### 11. `extension remove NAME`

```bash
bitbottle extension remove hello
bitbottle extension list | grep -E '^hello' || echo "removed"
```

Final line prints `removed`. The on-disk directory is gone.

### 12. Removing a missing extension errors clearly

```bash
bitbottle extension remove not-an-extension
```

Exit code: non-zero. stderr mentions the extension is not installed.

## Cleanup

```bash
bitbottle extension remove hello    2>/dev/null || true
bitbottle extension remove ext-noop 2>/dev/null || true
rm -rf /tmp/bitbottle-ext-noop
```
