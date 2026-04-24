# bitbottle

`bitbottle` is a command-line interface for self-hosted Bitbucket Server and Bitbucket Data Center. It follows the same design philosophy as [GitHub CLI](https://github.com/cli/cli): a factory-injected dependency model, TTY-aware output, and machine-readable non-TTY output for scripting.

## Installation

```bash
go install github.com/proggarapsody/bitbottle/cmd/bitbottle@latest
```

Or build from source:

```bash
git clone https://github.com/proggarapsody/bitbottle
cd bitbottle
make build
```

## Authentication

```bash
bitbottle auth login --hostname bitbucket.example.com
bitbottle auth status
```

## Usage

```bash
# List repositories (auto-detects host from config)
bitbottle repo list
bitbottle repo list --limit 10

# List pull requests for a repo (detects from git remote or explicit arg)
bitbottle pr list
bitbottle pr list MYPROJECT/my-service
bitbottle pr list --state merged
bitbottle pr list --state closed --limit 5

# Override host when multiple Bitbucket instances are configured
bitbottle repo list --hostname bitbucket.example.com
```

### TTY vs non-TTY output

When stdout is a terminal, `bitbottle` prints aligned columns with a header row:

```
SLUG           PROJECT  TYPE
my-service     MYPROJ   git
another-repo   MYPROJ   git
```

When piped or redirected, output is tab-separated with no headers — safe to parse with `awk`, `cut`, or `jq`:

```bash
bitbottle repo list | awk '{print $1}'
```

## Configuration

Configuration is stored at `~/.config/bitbottle/hosts.yml`:

```yaml
bitbucket.example.com:
  oauth_token: <token>
  git_protocol: ssh
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
