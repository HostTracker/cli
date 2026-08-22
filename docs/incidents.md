# `ht-cli incidents`

Downtime incidents, their checks and their comments

## `ht-cli incidents comment`

Annotate an incident and get the incident back.

```
ht-cli incidents comment <id> [flags]
```

Attaches a short comment to an incident and answers with the incident as it now stands, so no follow-up read is needed. Use it to record a cause or a resolution alongside the outage itself. Only the monitor's owner may comment - read access through a publicly shared monitor is not enough - and the write is convergent, so a retried submission changes nothing twice.

POST /monitor/incident/{id}/comment (commentIncident)

Arguments:
  <id>	the incident id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli incidents get`

Get one incident, with the transitions that opened and closed it.

```
ht-cli incidents get <id> [flags]
```

Returns a single incident including the transition that opened it, the transition that closed it if it is resolved, and the per-location verdicts recorded at each. Use it when an incident id is already in hand and the whole timeline is wanted; the list endpoint returns the same rows but is the wrong tool for one. The response carries the underlying check ids, so a result can be read directly from here. Each timeline entry also carries the assertion rules that failed and the policy codes that were violated, when the check recorded any. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it. expand=recheck adds the episode-level constellation beside the timeline.

GET /monitor/incident/{id} (getIncident)

Arguments:
  <id>	the incident id

| Flag | Type | Description |
|---|---|---|
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance \| recheck \| count) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitorId \| monitor \| start \| end \| durationSec \| state \| severity \| ... 6 more) (repeatable) |

## `ht-cli incidents list`

List down-episodes across the account or a monitor selection.

```
ht-cli incidents list [flags]
```

Returns a page of incidents - the episodes between a monitor going down and coming back - narrowed by monitor id (monitor=), address (url=, with like=true for substring match) or free text (q=), time window, severity or whether they are still open. The three monitor filters combine rather than adding to each other, exactly as they do on the monitor list, and with none of them this is the account's own incident feed over any window - an episode row exists per state change, so it stays small. Order it with sort=monitor to group the page by monitor (name A to Z, newest-first inside each) instead of the default sort=time. Each row carries its cause, duration, severity band and whether it fell inside a maintenance window. Use it instead of scanning raw results when the question is about outages rather than checks. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it. expand=recheck adds the constellation that opened each episode: the location that detected it, the locations that confirmed it grouped by the error each saw, and the locations that still saw the target up.

GET /monitor/incident (listIncident)

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/incident/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance \| recheck \| count) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitorId \| monitor \| start \| end \| durationSec \| state \| severity \| ... 6 more) (repeatable) |
| `--from` | int64 | The start of the time window, in Unix seconds. |
| `--like` | bool | true switches url= to substring match. |
| `--limit` | int64 | Rows to return. |
| `--monitor` | stringSlice | Monitor ids, ANY-OF - a NARROWING filter ANDed with url=/q=, and optional: omitting every monitor filter reads the whole account's incident feed... (repeatable) |
| `--q` | string | Free-text monitor filter: substring over name + address, the same matching GET /monitor's q= performs, ANDed with the other monitor filters. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/incident/q: inline JSON, @file, or - |
| `--severity` | stringSlice | minor \| major \| critical, ANY-OF. (minor \| major \| critical) (repeatable) |
| `--sort` | string | Which column to order by, optionally with a direction: sort=name or sort=name:desc. (time \| time:desc \| monitor \| monitor:asc \| monitor:desc) |
| `--state` | stringSlice | open \| resolved - ANY-OF. (open \| resolved) (repeatable) |
| `--to` | int64 | The end of the time window, in Unix seconds. |
| `--url` | stringSlice | Narrow by the monitor's ADDRESS instead of by GUID - exact (case-insensitive) unless like=true. (repeatable) |

## `ht-cli incidents list-check`

List the failing checks recorded inside one incident.

```
ht-cli incidents list-check <id> [flags]
```

Returns a page of the failed checks that occurred during one incident, each linking to its full result and page snapshot. Use it to see exactly what happened check by check during an outage, rather than re-deriving the episode from the raw result feed. Availability depends on the account's plan, and a plan refusal names the feature it needs. Each row carries the monitor's identifying projection and the recheck constellation behind the failure. Ask for expand=metrics to decode each check's stored measurements and, from the same document, the assertion rules it failed and the policy codes it violated - the per-check counterpart of the detail the incident's own timeline carries for the transitions that opened and closed the episode.

GET /monitor/incident/{id}/check (listIncidentCheck)

Arguments:
  <id>	the incident id

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/incident/{id}/check/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance \| metrics \| recheck) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitorId \| monitor \| at \| durationSec \| state \| checkNumber \| checkCount \| ... 10 more) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/incident/{id}/check/q: inline JSON, @file, or - |

## `ht-cli incidents list-monitor`

List one monitor's down-episodes.

```
ht-cli incidents list-monitor <monitor-id> [flags]
```

Returns a page of the monitor's incidents, newest first, with the same window, severity, state and expansion options as the collection read, and no sort parameter - a single monitor's page has nothing to group by. Use it when the monitor is already known; the collection read at /monitor/incident serves cross-monitor and account-wide questions. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it. expand=recheck adds the constellation that opened each episode.

GET /monitor/{monitorId}/incident (listMonitorIncident)

Arguments:
  <monitor-id>	the monitor id

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/{monitorId}/incident/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance \| recheck \| count) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitorId \| monitor \| start \| end \| durationSec \| state \| severity \| ... 6 more) (repeatable) |
| `--from` | int64 | The start of the time window, in Unix seconds. |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/{monitorId}/incident/q: inline JSON, @file, or - |
| `--severity` | stringSlice | minor \| major \| critical, ANY-OF. (minor \| major \| critical) (repeatable) |
| `--state` | stringSlice | open \| resolved - ANY-OF. (open \| resolved) (repeatable) |
| `--to` | int64 | The end of the time window, in Unix seconds. |

---

[Back to the index](README.md)
