# `ht results`

Individual check results and their summaries

## `ht results get-monitor`

Get one check result in full detail.

```
ht results get-monitor <monitor-id> <id> [flags]
```

Returns everything recorded for a single check: the verdict from each monitoring location, the timing breakdown, the decoded error, and any recheck that confirmed or refuted a failure. Use it once a result id is in hand; the list endpoint returns the same shape but is the wrong tool for one row. The response also says whether a page snapshot was captured for this check. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it. Ask for expand=metrics to decode the check's stored measurements and, from the same document, the assertion rules it failed and the policy codes it violated.

GET /monitor/{monitorId}/result/{id} (getMonitorResult)

Arguments:
  <monitor-id>	the monitor id
  <id>	the result id

| Flag | Type | Description |
|---|---|---|
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance \| metrics \| recheck \| count) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitorId \| monitor \| at \| durationSec \| state \| checkNumber \| checkCount \| ... 10 more) (repeatable) |

## `ht results get-monitor-snapshot`

Download the page snapshot captured for a check result.

```
ht results get-monitor-snapshot <monitor-id> <id> [flags]
```

Returns the stored page snapshot for one check as binary image data rather than base64 inside a JSON body, with validators and cache headers so it can be served from a cache or a content network. A snapshot never changes once captured, so a client may cache it indefinitely - send the ETag back as If-None-Match and an unchanged snapshot answers 304 with no body. HEAD is accepted on the same address for a metadata-only probe.

GET /monitor/{monitorId}/result/{id}/snapshot (getMonitorResultSnapshot)

Arguments:
  <monitor-id>	the monitor id
  <id>	the result id

| Flag | Type | Description |
|---|---|---|
| `--as` | string | See this operation's description for how this parameter narrows the result. |
| `--if-none-match` | string | The ETag a previous answer carried. |

## `ht results get-summary`

Get uptime, SLA and response-time figures across monitors.

```
ht results get-summary [flags]
```

Returns aggregated uptime and SLA figures for the monitors named in the request over a time window - as one aggregate per monitor, or as a bucketed series ready to chart. Optional timing metrics add response time and its phases, each point carrying the mean, the 95th percentile and how many checks it was drawn from. Use it instead of reading raw results whenever the question is about a period rather than a check. A window that would produce too many buckets is refused with the largest window that would fit, not silently truncated. Every row carries the seconds it was built from - up, down, total measured, and the maintenance time split into the part the monitor spent up and the part it spent down - together with the checks recorded, split the same way. Read the uptime percentage rather than dividing those seconds yourself: on the whole-window aggregate it comes from the stored daily statistics, which is the figure the rest of the product reports. Set groupBy=account to fold the rows into one per bucket across the whole selection - seconds and counts summed, the percentage recomputed from those sums rather than averaged across monitors, plus how many monitors it covers, whether it covered the account or only a named selection, and whether the timing figures were sampled. The monitor filter is optional in that mode; because the roll-up reads every monitor it covers, its window is capped at 30 days and a selection of more than a couple of thousand monitors is refused rather than sampled - an availability figure over some of the monitors would be silently wrong. Timing metrics take the opposite trade there: over many monitors they are drawn from a bounded sample and the row says so, instead of being refused. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it. An account roll-up row aggregates many monitors, so it has no monitor to embed and the monitor expansion is refused there.

GET /monitor/result/summary (getResultSummary)

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/result/summary/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--bucket` | string | none \| hour \| day \| week \| month. (none \| hour \| day \| week \| month) |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance \| incidentCounts \| count) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (monitorId \| monitor \| from \| to \| upSec \| downSec \| totalSec \| maintenance \| ... 11 more) (repeatable) |
| `--from` | int64 | The start of the time window, in Unix seconds. |
| `--group-by` | string | monitor (the default) - one row per monitor per bucket; account - one row per bucket, aggregated across every monitor the request covers. (monitor \| account) |
| `--limit` | int64 | Rows to return. |
| `--metrics` | stringSlice | Which timing metrics to include. (responseTime \| dns \| connect \| tls \| ttfb \| transfer) (repeatable) |
| `--monitor` | stringSlice | The monitors to summarise. (repeatable) (required) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/result/summary/q: inline JSON, @file, or - |
| `--sla` | float64 | An ad-hoc SLA target in percent, overriding each monitor's own slaTarget for this request only. |
| `--to` | int64 | The end of the time window, in Unix seconds. |

## `ht results list`

List raw check results across one or more monitors.

```
ht results list [flags]
```

Returns a page of individual check results, newest first, narrowed by time window, monitoring location, outcome (state=up|down) and by monitor. The monitor filters combine: monitor= (ids), url= (addresses, with like=true for substring match) and q= (free text over name and address) all narrow the same set, exactly as they do on the monitor list, and sending none reads the whole account's feed. Because an unfiltered read is the whole account, its time window is capped: an omitted from= takes that window and a wider explicit one is refused rather than quietly clipped - name the monitors with monitor= to read further back. Order it with sort=monitor to group the page by monitor (name A to Z, newest-first inside each) instead of the default sort=time. The summary endpoint answers aggregate uptime questions far more cheaply. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it. Ask for expand=metrics to decode the check's stored measurements and, from the same document, the assertion rules it failed and the policy codes it violated.

GET /monitor/result (listResult)

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/result/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance \| metrics \| recheck \| count) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitorId \| monitor \| at \| durationSec \| state \| checkNumber \| checkCount \| ... 10 more) (repeatable) |
| `--from` | int64 | Window start, Unix seconds. |
| `--like` | bool | true switches url= from exact match to case-insensitive SUBSTRING match. |
| `--limit` | int64 | Rows to return. |
| `--location` | stringSlice | Agent (location) ids - ANY-OF, like every other v2 list filter. (repeatable) |
| `--monitor` | stringSlice | The monitors to read, as an id list - a NARROWING filter, ANDed with url= and q= exactly as on GET /monitor. (repeatable) |
| `--q` | string | Free-text monitor filter: case-insensitive substring over the monitor's NAME and address - the same matching GET /monitor's own q= performs, and... |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/result/q: inline JSON, @file, or - |
| `--sort` | string | Which column to order by, optionally with a direction: sort=name or sort=name:desc. (time \| time:desc \| monitor \| monitor:asc \| monitor:desc) |
| `--state` | stringSlice | Which lifecycle states to include. (up \| down) (repeatable) |
| `--to` | int64 | Window end, Unix seconds. |
| `--url` | stringSlice | Narrow by the monitor's ADDRESS instead of by GUID - list-accepting (ANY-OF), matched against the monitor's address. (repeatable) |

## `ht results list-monitor`

List one monitor's raw check results.

```
ht results list-monitor <monitor-id> [flags]
```

Returns a page of individual check results for the monitor in the path, newest first, with the same window, location, outcome (state=up|down) and expansion options as the cross-monitor list. It takes no sort parameter - a single monitor's page has nothing to group by - and no window cap, the path being the scope. Use it when the monitor is already known; the collection read at /monitor/result serves multi-monitor questions in one call. Nothing is embedded by default: ask for expand=monitor for the monitor's identifying projection, and expand=monitor.settings / monitor.subscription / monitor.lastIncident / monitor.maintenance to embed the monitor's own blocks inside it. Ask for expand=metrics to decode the check's stored measurements and, from the same document, the assertion rules it failed and the policy codes it violated.

GET /monitor/{monitorId}/result (listMonitorResult)

Arguments:
  <monitor-id>	the monitor id

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/{monitorId}/result/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance \| metrics \| recheck \| count) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitorId \| monitor \| at \| durationSec \| state \| checkNumber \| checkCount \| ... 10 more) (repeatable) |
| `--from` | int64 | The start of the time window, in Unix seconds. |
| `--limit` | int64 | Rows to return. |
| `--location` | stringSlice | Agent (location) ids - ANY-OF. (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/{monitorId}/result/q: inline JSON, @file, or - |
| `--state` | stringSlice | Which lifecycle states to include. (up \| down) (repeatable) |
| `--to` | int64 | The end of the time window, in Unix seconds. |

---

[Back to the index](README.md)
