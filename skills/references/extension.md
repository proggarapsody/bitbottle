# extension

| Command | Description |
|---|---|
| `bitbottle extension install USER/REPO` | Install a bitbottle extension from GitHub |
| `bitbottle extension install --local PATH` | Install a local extension via symlink |
| `bitbottle extension list` | List installed extensions |
| `bitbottle extension upgrade NAME` | Upgrade a single installed extension to the latest release |
| `bitbottle extension upgrade --all` | Upgrade all installed extensions |
| `bitbottle extension remove NAME` | Remove an installed extension |
| `bitbottle extension exec NAME [args...]` | Execute an installed extension; strips `*KEYRING_PASSPHRASE*`/`*KEYRING_PASSWORD*` env vars and injects `BB_TOKEN` + `BITBOTTLE_VERSION` |
| `bitbottle extension scaffold NAME` | Generate a new bitbottle extension project from a template in `./bitbottle-NAME/` |
| `bitbottle extension scaffold NAME --lang bash` | Scaffold a bash extension (default `--lang go`) |
| `bitbottle extension scaffold NAME --dir PATH` | Scaffold into `PATH/bitbottle-NAME/` instead of cwd |
