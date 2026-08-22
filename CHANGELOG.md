# Changelog

All notable changes to `ht-cli` are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [semantic versioning](https://semver.org/spec/v2.0.0.html): a
renamed or removed command is a breaking change, a new one is not.

## [Unreleased]

### Added

- First release of `ht-cli`, the official HostTracker command-line client.
- 139 generated commands in 14 groups, one per operation of the published
  API2 v2 specification. Path parameters are positional arguments, query
  parameters are flags, and each `POST <path>/q` body query is folded into
  the read it belongs to as `--query` and `--query-file`.
- `ht-cli auth login|logout|status` and `ht-cli config get|set|list|path`,
  with several named profiles in a `0600` file under the OS configuration
  directory.
- Convenience verbs for the protocols that take more than one call:
  `ht-cli check run` (start an instant check and follow it), `ht-cli jobs
  wait` (poll a job to a terminal state), `ht-cli webhooks verify` (check
  a delivery signature offline).
- `ht-cli api` as the escape hatch for an address the generated commands
  do not model, riding the same token, retries and error mapping.
- `--output json|yaml|table`, defaulting to a table on a terminal and JSON
  when piped; `--all` to walk every page; `--verbose` to trace requests,
  request ids and the rate-limit budget on stderr.
- An exit-code table that separates usage, auth, not-found, validation,
  rate-limit and network failures.
- Generated command reference under `docs/`, and shell completion for
  bash, zsh, fish and PowerShell.
- `install.sh`, a POSIX shell installer that downloads the release
  archive for the running platform, checks it against the release
  `checksums.txt` and installs the binary into `~/.local/bin`.
- Packages on every release beside the archives: a `.deb`, `.rpm` and
  `.apk` per Linux architecture, each carrying the binary in `/usr/bin`
  and the bash, zsh and fish completions; and the multi-architecture
  image `ghcr.io/hosttracker/ht-cli` for CI jobs. Neither needs a
  credential of its own.
- Signed apt and dnf repositories at https://hosttracker.github.io/apt,
  so `apt install ht-cli` and `dnf install ht-cli` work after one setup
  command and `apt upgrade` carries new versions. The release workflow
  rebuilds and signs both from the packages of each release and pushes
  them to `HostTracker/apt`, which serves them through GitHub Pages.
- Package-manager channels, each published from the release workflow when
  its credential is configured and skipped when it is not: the Homebrew
  tap, the Scoop bucket `HostTracker/scoop-bucket`, a winget manifest
  pull request for `HostTracker.ht-cli`, and the AUR package
  `ht-cli-bin`.

### Changed

- Renamed the binary to `ht-cli`, before the first release. The short
  name `ht` is taken by other packages (the HT Editor formula in Homebrew
  core, the Debian and Arch `ht` package, a crates.io tool), so it would
  have collided on installation. Everything else is unchanged: the module
  stays `github.com/HostTracker/cli`, and the `HT_TOKEN`, `HT_BASE_URL`,
  `HT_PROFILE`, `HT_OUTPUT` and `HT_CONFIG_DIR` variables keep their
  names because they are HostTracker-wide. The configuration directory
  moves from `<config>/ht` to `<config>/ht-cli`, and the entry point
  moved to `cmd/ht-cli`, so that a `go install` of
  `github.com/HostTracker/cli/cmd/ht-cli@latest` installs a binary under
  the right name.
