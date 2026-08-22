# `ht maintenance`

Planned maintenance windows

## `ht maintenance create`

Schedule a maintenance window over an explicit set of monitors.

```
ht maintenance create [flags]
```

Creates a maintenance window and answers with it, including the coverage it resolved to. Use it to suppress alerts, and optionally statistics, for planned downtime - disabling the monitors instead would also erase the difference between planned and unplanned outage in the record. Name the monitors it covers in one of two ways, never both: monitorIds with a single suppress when they are all treated the same, or monitors with a suppression each when they are not. A window with neither is refused - a window covering nothing silences nothing. The window is anchored to the time zone supplied, an IANA id, so a recurring window keeps its local hour across a daylight-saving change.

POST /maintenance (createMaintenance)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht maintenance delete`

Cancel a maintenance window and receive a receipt.

```
ht maintenance delete <id> [flags]
```

Cancels a maintenance window immediately and answers with a receipt describing what was removed, including whether the window was actively suppressing alerts at the moment it was cancelled - which decides whether alerting resumes right now. The receipt is what confirms the cancellation, so no follow-up read is needed.

DELETE /maintenance/{id} (deleteMaintenance)

Arguments:
  <id>	the maintenance id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht maintenance get`

Read one maintenance window.

```
ht maintenance get <id> [flags]
```

Returns a single maintenance window in exactly the shape the list returns it in - which is what makes the Location header a created window answers with worth following, and what makes the row safe to send straight back to an update. A window that does not exist and one that belongs to another account answer the same 404. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it.

GET /maintenance/{id} (getMaintenance)

Arguments:
  <id>	the maintenance id

| Flag | Type | Description |
|---|---|---|
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| name \| from \| to \| durationSec \| timezone \| recurrence \| enabled \| ... 7 more) (repeatable) |

## `ht maintenance list`

List scheduled, active and finished maintenance windows.

```
ht maintenance list [flags]
```

Returns a page of maintenance windows, optionally narrowed by time window, state or the monitors they cover. A recurring window is returned once with its recurrence rule rather than expanded into every future occurrence, so a year of weekly maintenance is one row. The envelope carries a sync cursor, so a poll loop can read only what changed. Order it with sort=from (the default, newest window start first) or sort=created, each with an optional :asc/:desc suffix. Every row carries its coverage per monitor in monitors[]; the window-level suppress is present only when every monitor it covers shares one suppression. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it.

GET /maintenance (listMaintenance)

Filters too long for a query string go in a body query: --query-file filter.json (POST /maintenance/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | monitor - the identifying projection of the window's monitor set. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| name \| from \| to \| durationSec \| timezone \| recurrence \| enabled \| ... 7 more) (repeatable) |
| `--from` | int64 | Window START lower bound, Unix seconds. |
| `--limit` | int64 | Rows to return. |
| `--monitor` | stringSlice | Only windows covering these monitors. (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /maintenance/q: inline JSON, @file, or - |
| `--sort` | string | from (the default) \| created, optionally suffixed with :asc or :desc (sort=created:asc). (from \| created \| from:asc \| from:desc \| created:asc \| created:desc) |
| `--state` | stringSlice | scheduled \| active \| finished, any combination. (scheduled \| active \| finished) (repeatable) |
| `--to` | int64 | Window START upper bound, Unix seconds. |
| `--updated-since` | string | Return only what changed since this point - either Unix seconds, or the syncCursor a previous response returned. |

## `ht maintenance list-monitor`

List the maintenance windows covering one monitor.

```
ht maintenance list-monitor <monitor-id> [flags]
```

Returns the maintenance windows that cover the monitor in the path, each as its window DEFINITION in exactly the shape the maintenance list returns - a recurring window appears once, never expanded into occurrences. The monitor-side answer to the monitor-maintenance relationship; the same rows can be inlined on a monitor read with expand=maintenance, and the window-side direction is the maintenance list's expand=monitor. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it.

GET /monitor/{monitorId}/maintenance (listMonitorMaintenance)

Arguments:
  <monitor-id>	the monitor id

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/{monitorId}/maintenance/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| name \| from \| to \| durationSec \| timezone \| recurrence \| enabled \| ... 7 more) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/{monitorId}/maintenance/q: inline JSON, @file, or - |

## `ht maintenance update`

Reschedule a maintenance window or change what it covers.

```
ht maintenance update <id> [flags]
```

Applies a partial update to a maintenance window - its schedule, time zone, name, what it suppresses, or the monitors it covers - and answers with the window as it now stands. Members the body omits are unchanged. Use it rather than deleting and recreating when only part of the configuration moves. Coverage, when sent, REPLACES the previous coverage rather than adding to it: monitors states the whole set with a suppression each, monitorIds with suppress gives every listed monitor the same one, monitorIds alone keeps each already-covered monitor's own suppression, and suppress alone applies one to every monitor the window already covers. Adding a monitor with monitorIds alone to a window whose monitors do not all share one suppression is refused, because there would be nothing to give the newcomer.

PATCH /maintenance/{id} (updateMaintenance)

Arguments:
  <id>	the maintenance id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

---

[Back to the index](README.md)
