# `ht monitors`

Create, read, change and bulk-edit monitors

## `ht monitors bulk-create`

Create many monitors at once as one asynchronous job.

```
ht monitors bulk-create [flags]
```

Submits a batch of monitor definitions as a single job and answers immediately with a job id to poll for per-item results. Shared defaults are declared once and overridden per item. Prefer it over a loop of single creates when importing a site list; per-item validation happens as the job runs, so the response here only confirms the batch was accepted. An Idempotency-Key is mandatory, because a retried batch would otherwise duplicate every item in it. How many items one batch may carry is the account's own monitor cap, never more than 5000; the number is published as limits.maxBulkItems on the account endpoint, and a batch over it is refused naming the cap that applies. The 202 carries a Retry-After header saying how long to wait before the first poll, sized on how much work was accepted.

POST /monitor/bulk (bulkCreateMonitor)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required. A client-chosen key, unique per logical request. (required) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors bulk-delete`

Delete every monitor a filter selects, as an asynchronous job.

```
ht monitors bulk-delete [flags]
```

Resolves the filter to a concrete set of monitors, answers with a job id and the number accepted, and deletes them one at a time - each in its own transaction, each with its own receipt of what was cascaded away. The set is fixed at submission: a monitor that starts matching the filter afterwards is not touched. Check what a filter selects first with bulk-delete-validate, then send that filter here together with the count it reported as expectedCount: the server re-resolves and refuses if the number moved, so a delete can never be wider than the caller was shown. A filter that narrows by nothing is refused rather than taken to mean the whole account, and a selection larger than one operation carries is refused rather than trimmed. An Idempotency-Key is mandatory: this call answers 202 and then works asynchronously, so a retry after a timeout would otherwise start a second job over the same targets. The 202 carries a Retry-After header saying how long to wait before the first poll, sized on how much work was accepted.

POST /monitor/bulk-delete (bulkDeleteMonitor)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required. A client-chosen key, unique per logical request. (required) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors bulk-delete-validate`

Check which monitors a delete filter selects, without deleting anything.

```
ht monitors bulk-delete-validate [flags]
```

Resolves a filter and answers how many monitors it selects together with a sample of them, identified by name, url and type. Nothing is written and no selection token is minted. It is the verification step of a bulk delete: run it, show a human what is about to go, then send the same filter to the delete itself. The answer is a snapshot - monitors created or removed between the two calls change what the delete resolves, which is why the delete reports its own count as well.

POST /monitor/bulk-delete-validate (bulkDeleteValidateMonitor)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | replay key for this write (default: a fresh one per call) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors bulk-update`

Patch, or reset the statistics of, many monitors as one asynchronous job.

```
ht monitors bulk-update [flags]
```

Applies one partial patch, a statistics reset, or both to many monitors at once and answers with the job to poll. The target set is named either explicitly by ids or by a filter, and a filter that narrows by nothing is refused rather than taken to mean the whole account. Check what a filter selects, and that the patch parses, with bulk-update-validate first. Sending both a patch and a reset creates two jobs and the response names the second one as well. Tags can be edited as a DELTA - addTags and removeTags leave every other tag on each monitor in place, which a patch of tags (a replacement) cannot do across a set; the two spellings are mutually exclusive, and each item's receipt carries the tags it ended up with. An Idempotency-Key is mandatory: this call answers 202 and then works asynchronously, so a retry after a timeout would otherwise start a second job over the same targets. The 202 carries a Retry-After header saying how long to wait before the first poll, sized on how much work was accepted.

POST /monitor/bulk-update (bulkUpdateMonitor)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required. A client-chosen key, unique per logical request. (required) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors bulk-update-validate`

Check which monitors a bulk edit would touch, without changing anything.

```
ht monitors bulk-update-validate [flags]
```

Takes the same targets a bulk update takes - explicit ids or a filter - and answers how many monitors they select together with a sample of them, identified by name, url and type. The patch is parsed too, so a malformed change is reported here rather than as one failed item per monitor. Nothing is written. Run it before submitting a filter-driven edit, so the set being changed is one a human has seen.

POST /monitor/bulk-update-validate (bulkUpdateValidateMonitor)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | replay key for this write (default: a fresh one per call) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors bulk-validate`

Validate a batch of monitor definitions without creating anything.

```
ht monitors bulk-validate [flags]
```

Runs the validation a bulk create would run, over the same body, and answers immediately with one verdict per item: whether it would be accepted, the errors that would refuse it, and whether the account's package still has room for it. The package verdict is computed across the batch in order, so item 51 is judged against the room the first fifty would consume. Nothing is written and no job is created. It accepts exactly the batch size the create does - the account's own cap - so a batch that passes here is one the create will take. It requires the write scope because it evaluates a write.

POST /monitor/bulk-validate (bulkValidateMonitor)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | replay key for this write (default: a fresh one per call) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors copy`

Copy a monitor to one or more new addresses, with its alerting and coverage.

```
ht monitors copy <monitor-id> [flags]
```

Creates one new monitor per address in the body, each reproducing the source monitor's type, complete settings (including credentials, which a read never returns), interval or cron schedule, tags, locations, attached sub-checks and SLA target - plus, unless you switch them off, its alert subscriptions, its report subscriptions and its maintenance coverage. The copies start running even when the source is paused; send overrides with enabled false to stage them instead. overrides applies the same members a partial update takes to every copy at once, and its settings merge onto the source's rather than replacing them. Every address is validated - format, duplicates, locations, and the account's remaining room - BEFORE the first monitor is written, so a refusal leaves the account untouched. Up to ten addresses answer immediately with the created monitors; a longer list is accepted as a job to poll, and the two do exactly the same work - which is why an Idempotency-Key is required above ten addresses: a job works after the response, so a retry without one would copy the monitor a second time.

POST /monitor/{monitorId}/copy (copyMonitor)

Arguments:
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required for some requests, optional for the rest - see this operation's description for the condition. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors create`

Create a monitor, optionally with its contacts and subscriptions in the same call.

```
ht monitors create [flags]
```

Creates one monitor of the given type and, in the same request, can create or bind the contacts it alerts and wire their alert and report subscriptions - which is what makes a full site setup a single call rather than a scripted sequence. Send dryRun=true to validate and preview the resulting monitor and contact actions without writing anything. onOverlimit chooses what happens when the package has no room: fail (the default) refuses the write, disable creates the monitor disabled with a package-limit reason. An Idempotency-Key is required whenever the body carries inline contacts, because a retry would otherwise re-send paid confirmation messages. `interval` is optional: omit it and the monitor is created at the account's default cadence, or - for a type that publishes `fixedInterval` on GET /monitor/type, which the product schedules itself - at that pinned cadence. A value sent for a pinned type is still checked against the account's intervals and the type's floor.

POST /monitor (createMonitor)

| Flag | Type | Description |
|---|---|---|
| `--dry-run` | bool | Validate and report what would happen, writing nothing. |
| `--idempotency-key` | string | Required for some requests, optional for the rest - see this operation's description for the condition. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors delete`

Delete a monitor and receive a receipt of everything removed with it.

```
ht monitors delete <id> [flags]
```

Deletes one monitor and answers with a receipt naming the monitor and counting the alert, report and maintenance subscriptions that were cascaded away with it. Those counts exist nowhere else once the delete has run, which is why this returns a body rather than an empty response - the verification read a bodiless delete would force is exactly what the receipt removes.

DELETE /monitor/{id} (deleteMonitor)

Arguments:
  <id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht monitors get`

Retrieve one monitor, with its full configuration.

```
ht monitors get <id> [flags]
```

Returns a single monitor in exactly the row shape the list endpoint uses, but with settings embedded by default because this is the configuration read. Reach for it when the id is already known; filtering the list down to one id costs the same call and returns less. It takes the same expand vocabulary the list does, so lastResult, spans and uptime are available here too. A monitor that does not exist and one that belongs to another account answer the same not-found, so an id cannot be probed.

GET /monitor/{id} (getMonitor)

Arguments:
  <id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (settings \| attached \| subscription \| lastIncident \| maintenance \| lastResult \| lastResult.metrics \| lastResult.recheck \| ... 4 more) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| type \| name \| url \| effectiveUrl \| state \| since \| enabled \| ... 23 more) (repeatable) |
| `--from` | int64 | The start of the time window, in Unix seconds. |
| `--to` | int64 | The end of the time window, in Unix seconds. |

## `ht monitors get-attached`

Read the results of a monitor's attached sub-checks.

```
ht monitors get-attached <monitor-id> [flags]
```

Returns the current result of every sub-check attached to the monitor - blacklist, certificate expiry, domain expiry and web-risk - normalised into one shape, with null for a kind the monitor does not run. Use it for a focused read of sub-check state; embedding the same block into a monitor read is what the attached expand value is for. A monitor whose sub-checks have not run yet answers with null blocks, not with a not-found.

GET /monitor/{monitorId}/attached (getMonitorAttached)

Arguments:
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (dnsbl \| sslExp \| domainExp \| webRisk) (repeatable) |

## `ht monitors list`

List the account's monitors, filtered, sorted and cursor-paginated.

```
ht monitors list [flags]
```

Returns a page of monitors in a lean projection - identity, type, url, state, tags and the share and logging flags - so a dashboard read costs one call. Narrow the set with id, type, tag, state, enabled, openStat, preset or the free-text q filter, and use expand to embed settings, subscriptions, the last incident, the last check result, uptime, state spans or account-wide aggregates in the same response. Send includeId to keep named monitors in the answer whatever the filters say - what a dashboard needs to keep a live selection visible while the user narrows the list. For an incremental poll, send updatedSince with the syncCursor the previous response returned instead of re-reading the whole account. Order it with sort=name|state|type|interval|lastChange|url|tags|created, optionally suffixed :asc or :desc (sort=name:desc); without a suffix each column takes its natural direction - the time columns lastChange and created newest-first, everything else A to Z. There is no separate order parameter. Add pausedLast=true to push every monitor that is not actually being checked to the end, whatever the sort column is. expand=summary returns the ten largest domains in topDomains; the count is fixed.

GET /monitor (listMonitor)

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--enabled` | bool | The CONFIGURED flag, matching the resource's own enabled member. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (settings \| attached \| subscription \| lastIncident \| maintenance \| lastResult \| lastResult.metrics \| lastResult.recheck \| ... 4 more) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| type \| name \| url \| effectiveUrl \| state \| since \| enabled \| ... 23 more) (repeatable) |
| `--from` | int64 | Window start, Unix seconds - read by the INTERVAL-scoped expands. |
| `--id` | stringSlice | Explicit ids. An OMITTED list means NONE - never "all"; an omitted PARAMETER is no filter. (repeatable) |
| `--include-id` | stringSlice | Ids to include in the answer WHATEVER the other filters say - the "keep the rows I have selected visible" knob a dashboard needs when the user... (repeatable) |
| `--like` | bool | true switches url= from exact match to case-insensitive SUBSTRING match. |
| `--limit` | int64 | Rows to return. |
| `--open-stat` | bool | Whether the monitor's statistics are publicly shared - the row's own openStat member. |
| `--paused-last` | bool | Order every monitor that is NOT being checked (state: "paused" - switched off, or suspended by a package limit) AFTER every monitor that is, whatever... |
| `--preset` | stringSlice | Monitors built from a server-side settings PRESET - today the single value bl:ru, the Russian blacklist check. (bl:ru) (repeatable) |
| `--q` | string | Case-insensitive substring over name AND url - the single replacement for the search-like pair. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/q: inline JSON, @file, or - |
| `--sort` | string | name \| state \| type \| interval \| lastChange \| url \| tags \| created, optionally suffixed with :asc or :desc (sort=name:desc). (name \| state \| type \| interval \| lastChange \| url \| tags \| created \| ... 16 more) |
| `--state` | stringSlice | up \| down \| paused \| maintenance, any combination. (up \| down \| paused \| maintenance) (repeatable) |
| `--tag` | stringSlice | Tags, matched exactly (storage keeps them as one comma-separated column). (repeatable) |
| `--to` | int64 | Window end, Unix seconds. |
| `--type` | stringSlice | One or more of the 14 v2 type tokens. (http \| waterfall \| ping \| port \| domainExp \| sslExp \| dnsbl \| webRisk \| ... 6 more) (repeatable) |
| `--updated-since` | string | Return only what changed since this point - either Unix seconds, or the syncCursor a previous response returned. |
| `--url` | stringSlice | Address filter - the same spelling the result and incident reads take: list-accepting, matched against the monitor's address, exact... (repeatable) |

## `ht monitors list-span`

List one monitor's up/down state spans in a window.

```
ht monitors list-span <monitor-id> [flags]
```

Returns the monitor's contiguous up/down spans intersecting the requested window, oldest first - the same rows the spans expand value inlines on the monitor object, as their own read so span history can be fetched without re-reading the monitor. A span exists per state CHANGE, not per check, so the answer for any window is small and arrives as one closed page. Every down span carries the incident id that names it, its comment and the check numbers it opens and closes at, so a rendered bar can link straight to the episode behind it. The pause and maintenance overlays a bar chart also needs ride the monitor read's spans expand value, beside the spans themselves.

GET /monitor/{monitorId}/span (listMonitorSpan)

Arguments:
  <monitor-id>	the monitor id

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/{monitorId}/span/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (from \| to \| up \| eventCount \| incidentId \| comment \| firstCheckNumber \| lastCheckNumber) (repeatable) |
| `--from` | int64 | The start of the time window, in Unix seconds. |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/{monitorId}/span/q: inline JSON, @file, or - |
| `--to` | int64 | The end of the time window, in Unix seconds. |

## `ht monitors mute-dnsbl`

Mute or unmute named blacklist zones on a monitor's blacklist check.

```
ht monitors mute-dnsbl <monitor-id> [flags]
```

Adds or removes mute flags on specific blacklist zones for one monitor and answers with the monitor's updated blacklist block, so the effect is visible without a second read. Use it to silence a listing you have accepted rather than disabling the whole blacklist check. Muting is convergent, so an Idempotency-Key is honoured but never required.

POST /monitor/{monitorId}/attached/dnsbl/mute (muteMonitorDnsbl)

Arguments:
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors reset-stats`

Reset one monitor's accumulated uptime statistics.

```
ht monitors reset-stats <monitor-id> [flags]
```

Queues a job that clears the monitor's historical statistics and answers with the job to poll; the job's single result is the monitor as it stands afterwards. Use it for one monitor - the bulk update endpoint carries the same operation for many. The optional body may name a webhook to call on completion and nothing else. An Idempotency-Key is mandatory: this call answers 202 and then works asynchronously, so a retry after a timeout would otherwise start a second reset job. The 202 carries a Retry-After header saying how long to wait before the first poll, sized on how much work was accepted.

POST /monitor/{monitorId}/reset-stats (resetMonitorStats)

Arguments:
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required. A client-chosen key, unique per logical request. (required) |

## `ht monitors resolve-url`

Resolve a target address to the IPs it currently answers with.

```
ht monitors resolve-url [flags]
```

Looks up the addresses the url's host resolves to right now, which is what a DNS or TLS monitor's expected-address list is filled from. It resolves names only: nothing is connected to, no redirect is followed and no monitor type is guessed. A host with no address records answers an empty list, because that is the answer; a lookup that could not be completed answers 503 with Retry-After instead, so an empty list is never a lookup failure in disguise.

POST /monitor/resolve (resolveMonitorUrl)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | replay key for this write (default: a fresh one per call) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors update`

Partially update a monitor and get the updated resource back.

```
ht monitors update <id> [flags]
```

Applies a partial update: a member the body omits is left alone, and an explicit null clears an optional field. The response is the monitor exactly as a read would render it, so no follow-up fetch is needed. A monitor's type cannot change after creation - create a new monitor instead - and inline contacts are create-only, so wire additional contacts through the contact endpoints.

PATCH /monitor/{id} (updateMonitor)

Arguments:
  <id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht monitors validate-cron`

Check a cron schedule before saving it, and see when it would run.

```
ht monitors validate-cron [flags]
```

Answers whether a monitor write would accept this cron expression for this account and, when it would, the next five times it fires - which is what catches an expression that parses perfectly and does not mean what its author intended. A schedule that would be refused is still a 200: the answer carries valid=false and a reason (the expression cannot be parsed, it never runs, it runs more often than once a minute, or the account's package does not include cron scheduling). It reads the same validator and the same package limits the monitor write does, so a preview and a save cannot disagree.

POST /monitor/validate-cron (validateMonitorCron)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | replay key for this write (default: a fresh one per call) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

---

[Back to the index](README.md)
