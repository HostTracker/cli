# `ht reports`

Report subscriptions and generated reports

## `ht reports bulk-write-subscription`

Wire and unwire many report subscriptions in one transaction.

```
ht reports bulk-write-subscription [flags]
```

The report twin of the alert diff door: create[] and delete[] hold entries, each the cross product of its monitors, its contacts and its frequencies, applied in ONE transaction after the whole request validates. Set allMonitors (or allContacts) to mean every one on the account - one side only. Scheduled reports go to email contacts: an explicitly named contact of another type is refused naming it, and allContacts covers the account's email contacts. A frequency the package does not include is refused before anything is written. Both write scopes are required on every call, because the door writes on the monitor side and the contact side alike.

POST /report/bulk (bulkWriteReportSubscription)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht reports delete-contact`

Remove the report subscription between this contact and one monitor.

```
ht reports delete-contact <id> <monitor-id> [flags]
```

The contact-side mirror of deleteMonitorReport. 404 if the pair had no subscription.

DELETE /contact/{id}/report/{monitorId} (deleteContactReport)

Arguments:
  <id>	the contact id
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht reports delete-contact-reports`

Remove ALL of this contact's report subscriptions.

```
ht reports delete-contact-reports <id> [flags]
```

Removes every report subscription that would send to this contact, in one call.

DELETE /contact/{id}/report (deleteContactReports)

Arguments:
  <id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht reports delete-monitor`

Remove the report subscription between this monitor and one contact.

```
ht reports delete-monitor <monitor-id> <contact-id> [flags]
```

Removes all frequencies for the pair. 404 if the pair had no subscription.

DELETE /monitor/{monitorId}/report/{contactId} (deleteMonitorReport)

Arguments:
  <monitor-id>	the monitor id
  <contact-id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht reports delete-monitor-reports`

Remove ALL of this monitor's report subscriptions.

```
ht reports delete-monitor-reports <monitor-id> [flags]
```

Removes every report subscription attached to this monitor in one call.

DELETE /monitor/{monitorId}/report (deleteMonitorReports)

Arguments:
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht reports generate`

Request a report over a set of monitors and a time range.

```
ht reports generate [flags]
```

Submits a rendering request and answers with a job id while the document is produced in the background - rendering can take a while and the service sheds load under pressure, so an inline wait would be the wrong contract. Poll the job, or name a webhook to be called when it finishes. The monitor list is explicit and required, so a report can never quietly cover the whole account. An Idempotency-Key is mandatory: this call answers 202 and then renders in the background, so a retry after a timeout would otherwise queue a second rendering of the same report. The 202 carries a Retry-After header saying how long to wait before the first poll, sized on how much work was accepted.

POST /monitor/report (generateReport)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required. A client-chosen key, unique per logical request. (required) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht reports get`

Get a generated report's metadata.

```
ht reports get <id> [flags]
```

Returns what a report covers - its type, format, time range, monitors and expiry - without transferring the document itself. Use it to check that a report is still available, or to describe one in a list, before spending the bandwidth of a download. A report past its expiry answers not-found rather than a state value.

GET /monitor/report/{id} (getReport)

Arguments:
  <id>	the report id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| type \| format \| range \| monitorIds \| sections \| sizeBytes \| expiresAt \| ... 2 more) (repeatable) |

## `ht reports get-contact`

Read the report subscription between this contact and one monitor.

```
ht reports get-contact <id> <monitor-id> [flags]
```

The mirror of getMonitorReport - the same subscription addressed from the contact side.

GET /contact/{id}/report/{monitorId} (getContactReport)

Arguments:
  <id>	the contact id
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (monitor \| frequencies \| created) (repeatable) |

## `ht reports get-content`

Download a generated report's rendered document.

```
ht reports get-content <id>
```

Streams the report's bytes with a suggested filename, in the format the report was requested in. Use it after the metadata read confirms the report exists. A document that has aged out of the cache is re-rendered rather than refused, so a stored report id keeps working; a busy renderer answers with a retry hint instead of failing outright.

GET /monitor/report/{id}/content (getReportContent)

Arguments:
  <id>	the report id

## `ht reports get-monitor`

Read the report subscription between this monitor and one contact.

```
ht reports get-monitor <monitor-id> <contact-id> [flags]
```

Returns the report-frequency set this contact receives for this monitor, or 404 if none.The contact's address is included only when the token also carries a contact read scope - a monitor-scoped token sees which contact it is, not how to reach it.

GET /monitor/{monitorId}/report/{contactId} (getMonitorReport)

Arguments:
  <monitor-id>	the monitor id
  <contact-id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (contact \| frequencies \| created) (repeatable) |

## `ht reports get-subscription`

Read one report-subscription pair from the flat list, by its id.

```
ht reports get-subscription <id> [flags]
```

Returns the same row shape as listReportSubscription for exactly one (monitor, contact) pair, addressed by the id each list row carries. 404 for an id that does not decode, does not exist, or is not the caller's on either side.

GET /report/{id} (getReportSubscription)

Arguments:
  <id>	the report id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitor \| contact \| frequencies \| created) (repeatable) |

## `ht reports list-by-contact`

List every report subscription on the account, grouped by contact.

```
ht reports list-by-contact [flags]
```

The same report wiring as listReportSubscription, grouped one element per contact with its subscribed monitors nested underneath. The mirror of listReportByMonitor.

GET /report/by-contact (listReportByContact)

Filters too long for a query string go in a body query: --query-file filter.json (POST /report/by-contact/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--contact-confirmed` | bool | See this operation's description for how this parameter narrows the result. |
| `--contact-id` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--contact-q` | string | See this operation's description for how this parameter narrows the result. |
| `--contact-type` | stringSlice | See this operation's description for how this parameter narrows the result. (email \| sms \| voiceCall \| http \| telegram \| viber \| facebook \| googleChat \| ... 3 more) (repeatable) |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (contact \| subscriptions) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--monitor-id` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--monitor-like` | bool | See this operation's description for how this parameter narrows the result. |
| `--monitor-q` | string | See this operation's description for how this parameter narrows the result. |
| `--monitor-tag` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--monitor-type` | stringSlice | See this operation's description for how this parameter narrows the result. (http \| waterfall \| ping \| port \| domainExp \| sslExp \| dnsbl \| webRisk \| ... 6 more) (repeatable) |
| `--monitor-url` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /report/by-contact/q: inline JSON, @file, or - |

## `ht reports list-by-monitor`

List every report subscription on the account, grouped by monitor.

```
ht reports list-by-monitor [flags]
```

The same report wiring as listReportSubscription, grouped one element per monitor with its subscribed contacts nested underneath - so each monitor's identity is carried once rather than repeated for every contact. Takes the same filters as the flat list.

GET /report/by-monitor (listReportByMonitor)

Filters too long for a query string go in a body query: --query-file filter.json (POST /report/by-monitor/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--contact-confirmed` | bool | See this operation's description for how this parameter narrows the result. |
| `--contact-id` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--contact-q` | string | See this operation's description for how this parameter narrows the result. |
| `--contact-type` | stringSlice | See this operation's description for how this parameter narrows the result. (email \| sms \| voiceCall \| http \| telegram \| viber \| facebook \| googleChat \| ... 3 more) (repeatable) |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (monitor \| subscriptions) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--monitor-id` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--monitor-like` | bool | See this operation's description for how this parameter narrows the result. |
| `--monitor-q` | string | See this operation's description for how this parameter narrows the result. |
| `--monitor-tag` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--monitor-type` | stringSlice | See this operation's description for how this parameter narrows the result. (http \| waterfall \| ping \| port \| domainExp \| sslExp \| dnsbl \| webRisk \| ... 6 more) (repeatable) |
| `--monitor-url` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /report/by-monitor/q: inline JSON, @file, or - |

## `ht reports list-contact`

List the monitors that report to this contact, with each monitor's frequency set.

```
ht reports list-contact <id> [flags]
```

The mirror of listMonitorReport - what reports to this contact, one row per monitor.

GET /contact/{id}/report (listContactReport)

Arguments:
  <id>	the contact id

Filters too long for a query string go in a body query: --query-file filter.json (POST /contact/{id}/report/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (monitor \| frequencies \| created) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /contact/{id}/report/q: inline JSON, @file, or - |

## `ht reports list-monitor`

List the contacts this monitor reports to, with each contact's frequency set.

```
ht reports list-monitor <monitor-id> [flags]
```

Returns who receives reports for this monitor - one row per contact, each carrying its set of report frequencies (daily/weekly/monthly/quarterly/yearly). Reports are Email-only. Set a subscription with PUT /monitor/{monitorId}/report/{contactId}. The contact's address is included only when the token also carries a contact read scope - a monitor-scoped token sees which contact it is, not how to reach it.

GET /monitor/{monitorId}/report (listMonitorReport)

Arguments:
  <monitor-id>	the monitor id

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/{monitorId}/report/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (contact \| frequencies \| created) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/{monitorId}/report/q: inline JSON, @file, or - |

## `ht reports list-subscription`

List every report subscription on the account, flat.

```
ht reports list-subscription [flags]
```

Returns every (monitor, contact) report pair the account holds as one flat list - both sides identified on every row, with the pair's frequency set. Narrow by the entity-prefixed filters of either side: monitor.id, monitor.type, monitor.tag, monitor.url (+monitor.like), monitor.q, and contact.id, contact.type, contact.confirmed, contact.q. The report twin of the flat alert-subscription list.

GET /report (listReportSubscription)

Filters too long for a query string go in a body query: --query-file filter.json (POST /report/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--contact-confirmed` | bool | See this operation's description for how this parameter narrows the result. |
| `--contact-id` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--contact-q` | string | See this operation's description for how this parameter narrows the result. |
| `--contact-type` | stringSlice | See this operation's description for how this parameter narrows the result. (email \| sms \| voiceCall \| http \| telegram \| viber \| facebook \| googleChat \| ... 3 more) (repeatable) |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitor \| contact \| frequencies \| created) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--monitor-id` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--monitor-like` | bool | See this operation's description for how this parameter narrows the result. |
| `--monitor-q` | string | See this operation's description for how this parameter narrows the result. |
| `--monitor-tag` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--monitor-type` | stringSlice | See this operation's description for how this parameter narrows the result. (http \| waterfall \| ping \| port \| domainExp \| sslExp \| dnsbl \| webRisk \| ... 6 more) (repeatable) |
| `--monitor-url` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /report/q: inline JSON, @file, or - |

## `ht reports list-type`

List the report types, formats and schedules available.

```
ht reports list-type [flags]
```

Returns the catalogue of report types the account can generate, with the output formats, content sections and delivery frequencies each supports. Use it to build a report request form or to validate a type before generating - frequencies are published as words, so nothing here has to be decoded.

GET /report/type (listReportType)

Filters too long for a query string go in a body query: --query-file filter.json (POST /report/type/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (type \| label \| formats \| sections \| frequencies) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /report/type/q: inline JSON, @file, or - |

## `ht reports set-contact`

Set the frequency set for this contact-and-monitor pair.

```
ht reports set-contact <id> <monitor-id> [flags]
```

The contact-side mirror of setMonitorReport: idempotently sets which report frequencies this contact receives for the monitor in the path. `frequencies` is the EXACT desired set (at least one); reports are Email-only, so a non-Email contact is refused. Use DELETE to remove the subscription.

PUT /contact/{id}/report/{monitorId} (setContactReport)

Arguments:
  <id>	the contact id
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht reports set-monitor`

Set the report-frequency set for this monitor-and-contact pair.

```
ht reports set-monitor <monitor-id> <contact-id> [flags]
```

Idempotently sets which report frequencies this contact receives for this monitor. `frequencies` is the EXACT desired set (at least one); use DELETE to remove the subscription. Reports are Email-only - a non-Email contact is refused. Answers with the resulting subscription. The contact's address is included only when the token also carries a contact read scope - a monitor-scoped token sees which contact it is, not how to reach it.

PUT /monitor/{monitorId}/report/{contactId} (setMonitorReport)

Arguments:
  <monitor-id>	the monitor id
  <contact-id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

---

[Back to the index](README.md)
