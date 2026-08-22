# `ht`, the HostTracker command-line client

[![CI](https://github.com/HostTracker/cli/actions/workflows/ci.yml/badge.svg)](https://github.com/HostTracker/cli/actions/workflows/ci.yml)
[![Go reference](https://pkg.go.dev/badge/github.com/HostTracker/cli.svg)](https://pkg.go.dev/github.com/HostTracker/cli)
[![MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`ht` drives the [HostTracker API2 v2 REST surface](https://www.host-tracker.com/apidocs/v2)
from a shell: monitors, contacts, alerts, reports, incidents, maintenance
windows, status pages, webhooks and on-demand checks.

Every operation of the API is a command. They are generated from the
published OpenAPI document, so the CLI covers the whole surface and cannot
drift behind it: **139 commands in 14 groups**, plus a hand-written layer
for authentication, output, paging and the protocols that take more than
one call (job waits, instant checks, webhook signatures).

Underneath it is the [Go SDK](https://github.com/HostTracker/hosttracker-sdk-go),
so the bearer token, the automatic `Idempotency-Key` on writes, the retry
ladder for `429 rate_limited` and `503`, and the RFC 9457 problem mapping
are the SDK's behaviour, not a second implementation of it.

```console
$ ht monitors list --state down --limit 3
ID                                    NAME          TYPE  STATE  URL
0192f3c1-6d0a-7b41-9c22-5b1f4a0e77aa  api gateway   http  down   https://api.example.com/health
0192f3c1-7f2b-7c58-8d31-2ac9e5b31c04  checkout      http  down   https://example.com/checkout
0192f3c1-9a44-7d6e-b0a7-71d2c8f9e155  eu edge       ping  down   edge-eu.example.com

3 shown, 3 matched
```

## Install

**Homebrew** (macOS and Linux):

```sh
brew install HostTracker/tap/ht
```

**Go** (1.24 or newer):

```sh
go install github.com/HostTracker/cli@latest
```

**Release binaries**: download the archive for your platform from the
[releases page](https://github.com/HostTracker/cli/releases), verify it
against `checksums.txt`, and put `ht` on your `PATH`.

```sh
curl -fsSLO https://github.com/HostTracker/cli/releases/latest/download/ht_1.0.0_linux_amd64.tar.gz
tar xzf ht_1.0.0_linux_amd64.tar.gz
sudo install ht /usr/local/bin/ht
```

Shell completion is installed with the Homebrew formula; otherwise
`ht completion bash|zsh|fish|powershell` prints the script.

## Authenticate

Mint a token on the HostTracker profile page, then:

```sh
ht auth login                    # prompts, verifies, and stores it 0600
ht auth status                   # which credential is in force, and whether it works
```

The token can also ride the command line or the environment, which is what
a CI job wants:

```sh
export HT_TOKEN=...
ht monitors list

ht --token "$HT_TOKEN" monitors list
```

Settings live in a YAML file under the OS configuration directory
(`ht config path` prints it), written `0600` because it holds tokens.
Several named profiles can coexist:

```sh
ht auth login --profile staging
ht --profile staging monitors list
ht config set output json
ht config list
```

Resolution order for every setting: the flag, then the environment
(`HT_TOKEN`, `HT_BASE_URL`, `HT_PROFILE`, `HT_OUTPUT`), then the profile,
then the default. `HT_CONFIG_DIR` moves the file itself, which is the
clean way to give a CI job a throwaway profile.

## Ten commands to start with

```sh
# 1. What the account looks like, and what it has left.
ht account get
ht account get-quota

# 2. Every monitor, walking all the pages.
ht monitors list --all

# 3. Narrow it: the down http monitors tagged prod.
ht monitors list --state down --type http --tag prod

# 4. One monitor in full, with its settings and last result embedded.
ht monitors get <monitor-id> --expand settings,lastResult

# 5. Create one. --set writes into the body at a dotted path.
ht monitors create --set name="api health" --set type=http \
                   --set url=https://api.example.com/health --set interval=5

# 6. Change it, or take it out of service.
ht monitors update <monitor-id> --json '{"enabled":false}'

# 7. A one-off check from named countries, followed to its result.
ht check run https://www.host-tracker.com --type http --country de --country us --wait

# 8. Uptime and response times over a window.
ht results get-summary --from 1735689600 --to 1738368000

# 9. A bulk edit is asynchronous: it answers with a job id to wait on.
job=$(ht monitors bulk-update --json @edit.json -o json | jq -r .jobId)
ht jobs wait "$job"

# 10. Prove a webhook delivery came from HostTracker, offline.
ht webhooks verify --secret "$HT_WEBHOOK_SECRET" --headers-file headers.txt < body.json
```

`ht api` is the escape hatch for anything the generated commands do not
reach, and the shortest way to reproduce a call from the reference:

```sh
ht api GET /monitor --query limit=5 --query state=down
ht api POST /contact --set type=email --set address=ops@example.com
```

## Command tree

| Group | What it covers | Commands |
|---|---|---|
| `ht monitors` | Create, read, change and bulk-edit monitors | 18 |
| `ht monitor-types` | The catalogue of check types and their settings schema | 3 |
| `ht results` | Individual check results and their summaries | 5 |
| `ht incidents` | Downtime incidents, their checks and their comments | 5 |
| `ht maintenance` | Planned maintenance windows | 6 |
| `ht contacts` | Notification contacts, contact groups and confirmations | 19 |
| `ht alerts` | Who is alerted about which monitor, and what was sent | 21 |
| `ht reports` | Report subscriptions and generated reports | 19 |
| `ht webhooks` | Webhook endpoints, their deliveries and test sends | 8 |
| `ht status-pages` | Public status pages, their incidents, templates and subscribers | 18 |
| `ht account` | The account, its quota and its usage | 5 |
| `ht monitoring-locations` | The monitoring locations checks can run from | 3 |
| `ht instant-checks` | One-off checks run on demand, without a monitor | 5 |
| `ht jobs` | Long-running batch jobs started by the bulk operations | 4 |

Plus the hand-written `ht auth`, `ht config`, `ht check run`,
`ht jobs wait`, `ht webhooks verify`, `ht api`, `ht version` and
`ht completion`. The full reference is in [`docs/`](docs/README.md), and
`ht <group> --help` prints the same thing at the terminal.

Command names follow the operation they stand for: the verb, then what
qualifies it once the group's own noun is dropped. `listMonitor` is
`ht monitors list`, `bulkCreateMonitor` is `ht monitors bulk-create`,
`getMonitorAttached` is `ht monitors get-attached`. Path parameters are
positional arguments; query parameters are flags.

## Filters, bodies and paging

**Query parameters are flags**, spelled in kebab-case. A list-valued one
is repeatable and also accepts a comma-separated value:

```sh
ht monitors list --state up --state down
ht monitors list --state up,down
```

**A filter too long or too structured for a query string** goes in a body
query, which the API exposes as `POST <path>/q`. It is folded into the
same command:

```sh
ht monitors list --query-file filter.json
ht monitors list --query '{"id":["...","..."]}'
```

**Request bodies** come from `--json` or are assembled with `--set`:

```sh
ht monitors create --json '{"name":"api","type":"http","url":"https://example.com"}'
ht monitors create --json @monitor.json
cat monitor.json | ht monitors create --json -
ht monitors create --set type=http --set url=https://example.com \
                   --set settings.timeout=30 --set tags.0=prod
```

A `--set` value is read as JSON when it parses as one, so `enabled=true`
is a boolean and `interval=5` a number; quote it (`name='"5"'`) to force a
string.

**Paging** returns one page by default. `--all` follows the cursor to
exhaustion and prints one merged answer:

```sh
ht monitors list --limit 200 --all
ht monitors list --limit 200 --cursor "$cursor"
```

**Writes carry an `Idempotency-Key`** automatically, which is what makes
the SDK's retry replay a stored answer instead of writing twice. Pass your
own with `--idempotency-key` when the key must survive a process restart.

## Output

`--output json|yaml|table`, or `-o`. The default is a table on a terminal
and JSON when the answer is piped, so `ht monitors list | jq` needs no
flag and neither does reading it yourself.

```sh
ht monitors list -o json | jq -r '.data[] | select(.state=="down") | .url'
ht account get -o yaml
```

The table view is a reading aid: it shows the scalar members of each row,
folds a nested one into `{n}` or `[n]`, renders the API's Unix seconds as
UTC instants, and says how many rows matched. Use JSON for anything a
script depends on.

`--verbose` reports each request, its status, its `X-Request-Id` and the
rate-limit budget on stderr, retries included, which is the first thing to
turn on when an answer surprises you.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The command did what it was asked. |
| 1 | A failure with no more specific code. |
| 2 | The command line was wrong: an unknown flag, a missing argument, a bad value. |
| 3 | The credential is missing, rejected or under-scoped. |
| 4 | The address names nothing. |
| 5 | The API refused the request: validation, a conflict, a precondition. |
| 6 | Throttled, or the quota is exhausted. |
| 7 | The API could not be reached, or faulted. |

`429` splits across two codes on purpose: `rate_limited` is worth retrying
and `quota_exceeded` is not, and only the problem code tells them apart.

A failure prints the RFC 9457 problem to stderr, with the machine code,
the human detail, the offending members and the request id to quote in a
support request. Under `-o json` the whole problem document is printed
instead, so a script can read it:

```console
$ ht monitors get 00000000-0000-4000-8000-000000000099
ht: not_found (HTTP 404)
    No such monitor.
    /id: value=00000000-0000-4000-8000-000000000099
    request 9515cf81-e64f-459d-b4e5-7d7d9f77ded2
    https://api2.host-tracker.com/problems/not_found
$ echo $?
4
```

## Regenerating

The commands under `cmd/gen` are generated from the published
specification and committed, so building the CLI never needs the document
and a spec change arrives as a reviewable diff of commands.

```sh
./scripts/regen.sh                                  # fetch the published spec
HT_SPEC=../openapi/openapi-3.0.json ./scripts/regen.sh
```

It runs `go run ./internal/gen`, rewrites `docs/` with `ht docs`, and
builds. A push to the specification repository also fires the
`regenerate` workflow here, which opens a pull request when the output
moved.

## Contributing

`go test ./...` runs everything; `go test ./internal/gen -update` rewrites
the generator's golden files after a deliberate change to the naming or
flag mapping. Keep `gofmt -l .` empty and `go vet ./...` quiet, which is
what CI checks.

A read-only smoke against a real account is opt-in and needs a token:

```sh
HT_TOKEN=... go test -tags live ./cmd -run TestLive -v
```

## License

MIT. See [LICENSE](LICENSE).
