# `ht` command reference

Generated with `ht docs`. Do not edit by hand.

ht is the official command-line client for the HostTracker API.

Every operation of the API is a command, grouped by the family it belongs
to: ht monitors list, ht contacts create, ht status-pages get. Reads print
a table on a terminal and JSON when piped; --output picks explicitly.

Start with:

  ht auth login                      store an API token
  ht monitors list                   the account's monitors
  ht check run https://example.com   a one-off check, no monitor needed

Tokens are minted on the HostTracker profile page.

## Global flags

| Flag | Type | Description |
|---|---|---|
| `--all` | bool | walk every page of a collection instead of the first one |
| `--base-url` | string | API root (default: https://api2.host-tracker.com) |
| `--no-retry` | bool | do not retry a throttled or unavailable answer |
| `--profile` | string | configuration profile to use (default: the current one) |
| `--timeout` | duration | per-attempt request deadline (default 30s) |
| `--token` | string | API token, overriding the profile and HT_TOKEN |
| `--verbose` | bool | report every request, its status, request id and rate-limit budget on stderr |
| `-o, --output` | string | output format: json, yaml, table (default: table on a terminal, json when piped) |

## Commands

| Group | What it covers | Commands |
|---|---|---|
| [`ht account`](account.md) | The account, its quota and its usage | 5 |
| [`ht alerts`](alerts.md) | Who is alerted about which monitor, and what was sent | 21 |
| [`ht api`](api.md) | Send one request to any API address | 1 |
| [`ht auth`](auth.md) | Store, inspect and remove the API token | 3 |
| [`ht check`](check.md) | Run a one-off check without creating a monitor | 1 |
| [`ht config`](config.md) | Read and write the stored settings | 4 |
| [`ht contacts`](contacts.md) | Notification contacts, contact groups and confirmations | 19 |
| [`ht incidents`](incidents.md) | Downtime incidents, their checks and their comments | 5 |
| [`ht instant-checks`](instant-checks.md) | One-off checks run on demand, without a monitor | 5 |
| [`ht jobs`](jobs.md) | Long-running batch jobs started by the bulk operations | 5 |
| [`ht maintenance`](maintenance.md) | Planned maintenance windows | 6 |
| [`ht monitor-types`](monitor-types.md) | The catalogue of check types and their settings schema | 3 |
| [`ht monitoring-locations`](monitoring-locations.md) | The monitoring locations checks can run from | 3 |
| [`ht monitors`](monitors.md) | Create, read, change and bulk-edit monitors | 18 |
| [`ht reports`](reports.md) | Report subscriptions and generated reports | 19 |
| [`ht results`](results.md) | Individual check results and their summaries | 5 |
| [`ht status-pages`](status-pages.md) | Public status pages, their incidents, templates and subscribers | 18 |
| [`ht version`](version.md) | Print the version of ht, its SDK and its toolchain | 1 |
| [`ht webhooks`](webhooks.md) | Webhook endpoints, their deliveries and test sends | 9 |

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
