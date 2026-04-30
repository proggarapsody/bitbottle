# Scenario: shell completion

**Backend:** N/A (does not contact a server).

Verifies that `completion` emits sourceable shell completion for each
supported shell.

## Steps

### 1. bash

```bash
bitbottle completion bash | head -3
bitbottle completion bash | bash -n
```

First command's output begins with a bash completion preamble (e.g.
`# bash completion …` / `_bitbottle()` / similar). The second command
performs a syntax check and must exit `0`.

### 2. zsh

```bash
bitbottle completion zsh | head -3
bitbottle completion zsh | zsh -n
```

Output starts with `#compdef bitbottle` (or equivalent). `zsh -n` exits `0`.

### 3. fish

```bash
bitbottle completion fish | head -3
```

Output contains `complete -c bitbottle …` lines. Exit code: `0`.
If `fish` is installed, additionally:

```bash
bitbottle completion fish | fish -n
```

…must exit `0`.

### 4. powershell

```bash
bitbottle completion powershell | head -3
```

Output begins with a PowerShell completion stub (e.g.
`Register-ArgumentCompleter` or `using namespace …`). Exit code: `0`.

### 5. Unknown shell

```bash
bitbottle completion tcsh
```

Exit code: non-zero. stderr names the unsupported shell and lists the
supported ones.

### 6. Source it once and use it

In a fresh interactive shell:

```bash
source <(bitbottle completion bash)   # or zsh equivalent
bitbottle p<TAB>
```

Tab completes to a `bitbottle` subcommand starting with `p` (e.g. `pr`,
`pipeline`).

## Cleanup

None.
