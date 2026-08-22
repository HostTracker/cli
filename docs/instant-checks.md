# `ht-cli instant-checks`

One-off checks run on demand, without a monitor

## `ht-cli instant-checks create`

Start a one-off check against a url or host.

```
ht-cli instant-checks create [flags]
```

Starts a check that runs immediately rather than on a schedule and answers with the identifiers to poll for its result, plus how long to wait before the first poll. Use it to test a url, verify a fix or run a diagnostic without creating a permanent monitor. When the checking pipeline is unavailable the request is refused outright - a check that was not registered never happened, and answering as if it had would hand back an id that never resolves.

POST /check (createInstantCheck)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli instant-checks get`

Fetch the state and result of an instant check.

```
ht-cli instant-checks get <db-id> <id> [flags]
```

Returns an instant check's current state and, once it has finished, its per-location results. While it is still running the response says how long to wait before asking again, both in the body and as a header, so a client never has to guess a polling interval. Use the history endpoint to browse past checks rather than this one.

GET /check/{dbId}/{id} (getInstantCheck)

Arguments:
  <db-id>	the federation the check was registered in
  <id>	the check id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| dbId \| state \| url \| type \| created \| doneAt \| retryAfter \| ... 1 more) (repeatable) |

## `ht-cli instant-checks list`

List the account's past instant checks.

```
ht-cli instant-checks list [flags]
```

Returns a page of the instant checks this account has run, newest first, optionally narrowed by time window or check type. Use it to find or audit a past ad-hoc check; to read one check's outcome, address it directly by the identifiers the start call returned.

GET /check (listInstantCheck)

Filters too long for a query string go in a body query: --query-file filter.json (POST /check/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| dbId \| type \| url \| state \| up \| created \| doneAt) (repeatable) |
| `--from` | int64 | The start of the time window, in Unix seconds. |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /check/q: inline JSON, @file, or - |
| `--to` | int64 | The end of the time window, in Unix seconds. |
| `--type` | stringSlice | v2 type tokens. Repeatable and comma-separable. (crawl \| dns \| dnsbl \| http \| ping \| port \| rusRegBL \| trace \| ... 3 more) (repeatable) |

## `ht-cli instant-checks list-device`

List the device profiles a page-loading check can emulate.

```
ht-cli instant-checks list-device [flags]
```

Returns the device profiles a waterfall check can be run as - the browser is emulated as that device, which changes the viewport, the pixel density and the user agent the target sees. A row's device is the value the deviceEmulation setting takes, on an instant check and on a monitor alike. Rows arrive in the order a picker should offer them. The same names ride the waterfall row of the instant-check type catalogue; read this endpoint when the profiles themselves are what you are after.

GET /check/device (listInstantCheckDevice)

Filters too long for a query string go in a body query: --query-file filter.json (POST /check/device/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (device \| priority) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /check/device/q: inline JSON, @file, or - |

## `ht-cli instant-checks list-type`

List the instant-check types and their options.

```
ht-cli instant-checks list-type [flags]
```

Returns the catalogue of instant-check types with a description, an example target and any per-type options such as device profiles or record types. Use it to discover what can be checked, and with which options, before starting a check.

GET /check/type (listInstantCheckType)

Filters too long for a query string go in a body query: --query-file filter.json (POST /check/type/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (type \| label \| description \| example \| experimental \| agentRouted \| retryAfter \| estimatedDurationSec \| ... 1 more) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /check/type/q: inline JSON, @file, or - |

---

[Back to the index](README.md)
