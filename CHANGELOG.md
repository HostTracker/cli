# Changelog

All notable changes to `ht` are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [semantic versioning](https://semver.org/spec/v2.0.0.html): a
renamed or removed command is a breaking change, a new one is not.

## [Unreleased]

### Added

- First release of `ht`, the official HostTracker command-line client.
- 139 generated commands in 14 groups, one per operation of the published
  API2 v2 specification. Path parameters are positional arguments, query
  parameters are flags, and each `POST <path>/q` body query is folded into
  the read it belongs to as `--query` and `--query-file`.
- `ht auth login|logout|status` and `ht config get|set|list|path`, with
  several named profiles in a `0600` file under the OS configuration
  directory.
- Convenience verbs for the protocols that take more than one call:
  `ht check run` (start an instant check and follow it), `ht jobs wait`
  (poll a job to a terminal state), `ht webhooks verify` (check a delivery
  signature offline).
- `ht api` as the escape hatch for an address the generated commands do
  not model, riding the same token, retries and error mapping.
- `--output json|yaml|table`, defaulting to a table on a terminal and JSON
  when piped; `--all` to walk every page; `--verbose` to trace requests,
  request ids and the rate-limit budget on stderr.
- An exit-code table that separates usage, auth, not-found, validation,
  rate-limit and network failures.
- Generated command reference under `docs/`, and shell completion for
  bash, zsh, fish and PowerShell.
