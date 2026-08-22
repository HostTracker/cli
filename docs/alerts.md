# `ht-cli alerts`

Who is alerted about which monitor, and what was sent

## `ht-cli alerts bulk-write-subscription`

Wire and unwire many alert subscriptions in one transaction.

```
ht-cli alerts bulk-write-subscription [flags]
```

Applies a whole subscription change at once: create[] and delete[] each hold entries, and an entry is the cross product of its monitors, its contacts and its alert types. Set allMonitors (or allContacts) to let the entries that omit that list mean every one on the account - only one of the two sides may be a wildcard, and an entry may not name a list the wildcard already covers. Use it instead of the per-pair PUT whenever more than one pair changes: everything is validated first and applied in ONE transaction, so a partial result is impossible, and an id the account does not own is refused naming the entry it came from. The answer says what actually changed. Both write scopes are required on every call, because the door writes on the monitor side and the contact side alike.

POST /alert/bulk (bulkWriteAlertSubscription)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli alerts delete-contact`

Remove the alert subscription between this contact and one monitor.

```
ht-cli alerts delete-contact <id> <monitor-id> [flags]
```

The contact-side mirror of deleteMonitorAlert. 404 if the pair had no subscription.

DELETE /contact/{id}/alert/{monitorId} (deleteContactAlert)

Arguments:
  <id>	the contact id
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli alerts delete-contact-alerts`

Remove ALL of this contact's alert subscriptions.

```
ht-cli alerts delete-contact-alerts <id> [flags]
```

Removes every alert subscription that would notify this contact, in one call.

DELETE /contact/{id}/alert (deleteContactAlerts)

Arguments:
  <id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli alerts delete-monitor`

Remove the alert subscription between this monitor and one contact.

```
ht-cli alerts delete-monitor <monitor-id> <contact-id> [flags]
```

Removes all alert types for the pair. 404 if the pair had no subscription.

DELETE /monitor/{monitorId}/alert/{contactId} (deleteMonitorAlert)

Arguments:
  <monitor-id>	the monitor id
  <contact-id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli alerts delete-monitor-alerts`

Remove ALL of this monitor's alert subscriptions.

```
ht-cli alerts delete-monitor-alerts <monitor-id> [flags]
```

Removes every alert subscription attached to this monitor in one call.

DELETE /monitor/{monitorId}/alert (deleteMonitorAlerts)

Arguments:
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli alerts get-contact`

Read the alert subscription between this contact and one monitor.

```
ht-cli alerts get-contact <id> <monitor-id> [flags]
```

The mirror of getMonitorAlert - the same subscription addressed from the contact side.

GET /contact/{id}/alert/{monitorId} (getContactAlert)

Arguments:
  <id>	the contact id
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (monitor \| alertTypes \| created) (repeatable) |

## `ht-cli alerts get-monitor`

Read the alert subscription between this monitor and one contact.

```
ht-cli alerts get-monitor <monitor-id> <contact-id> [flags]
```

Returns the alert-type set this contact receives for this monitor, or 404 if the pair has none.The contact's address is included only when the token also carries a contact read scope - a monitor-scoped token sees which contact it is, not how to reach it.

GET /monitor/{monitorId}/alert/{contactId} (getMonitorAlert)

Arguments:
  <monitor-id>	the monitor id
  <contact-id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (contact \| alertTypes \| created) (repeatable) |

## `ht-cli alerts get-notification`

Get one delivered notification, including its rendered content.

```
ht-cli alerts get-notification <id> [flags]
```

Returns the full record of one delivered notification: the rendered subject and body as the recipient saw them, plus every delivery attempt logged against it. Use it after finding an id through the notification list, when the summary fields are not enough to explain what a recipient received.

GET /contact/notification/{id} (getNotification)

Arguments:
  <id>	the notification id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| sentAt \| kind \| channel \| gateway \| contact \| monitor \| checkNumber \| ... 4 more) (repeatable) |

## `ht-cli alerts get-notification-summary`

Get per-contact delivery counts by outcome and day.

```
ht-cli alerts get-notification-summary [flags]
```

Returns one row per (contact, delivery outcome, UTC day) with the count of notifications delivered in that cell, over the requested window (the last month when no window is given). Use it to chart delivery volume or spot a silently failing channel without walking the whole log; outcomes use the same vocabulary the log's outcome filter takes.

GET /contact/notification/summary (getNotificationSummary)

Filters too long for a query string go in a body query: --query-file filter.json (POST /contact/notification/summary/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--contact` | stringSlice | Which contacts to include. (repeatable) |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (contact \| outcome \| day \| count) (repeatable) |
| `--from` | int64 | The start of the time window, in Unix seconds. |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /contact/notification/summary/q: inline JSON, @file, or - |
| `--to` | int64 | The end of the time window, in Unix seconds. |

## `ht-cli alerts get-subscription`

Read one alert-subscription pair from the flat list, by its id.

```
ht-cli alerts get-subscription <id> [flags]
```

Returns the same row shape as listAlertSubscription for exactly one (monitor, contact) pair, addressed by the id each list row carries. 404 for an id that does not decode, does not exist, or is not the caller's on either side.

GET /alert/{id} (getAlertSubscription)

Arguments:
  <id>	the alert id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitor \| contact \| alertTypes \| created) (repeatable) |

## `ht-cli alerts list-by-contact`

List every alert subscription on the account, grouped by contact.

```
ht-cli alerts list-by-contact [flags]
```

The same alert wiring as listAlertSubscription, grouped one element per contact with its subscribed monitors nested underneath. Takes the same filters as the flat list; the mirror of listAlertByMonitor.

GET /alert/by-contact (listAlertByContact)

Filters too long for a query string go in a body query: --query-file filter.json (POST /alert/by-contact/q).

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
| `--query` | string | filter with a body query against /alert/by-contact/q: inline JSON, @file, or - |

## `ht-cli alerts list-by-monitor`

List every alert subscription on the account, grouped by monitor.

```
ht-cli alerts list-by-monitor [flags]
```

The same alert wiring as listAlertSubscription, grouped one element per monitor with its subscribed contacts nested underneath - so each monitor's identity is carried once rather than repeated for every contact. Takes the same filters as the flat list; reach for it to render a monitor-first subscription matrix without de-duplicating the monitor rows yourself.

GET /alert/by-monitor (listAlertByMonitor)

Filters too long for a query string go in a body query: --query-file filter.json (POST /alert/by-monitor/q).

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
| `--query` | string | filter with a body query against /alert/by-monitor/q: inline JSON, @file, or - |

## `ht-cli alerts list-contact`

List the monitors that alert this contact, with each monitor's alert-type set.

```
ht-cli alerts list-contact <id> [flags]
```

The mirror of listMonitorAlert - what alerts this contact, one row per monitor.

GET /contact/{id}/alert (listContactAlert)

Arguments:
  <id>	the contact id

Filters too long for a query string go in a body query: --query-file filter.json (POST /contact/{id}/alert/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (monitor \| alertTypes \| created) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /contact/{id}/alert/q: inline JSON, @file, or - |

## `ht-cli alerts list-contact-notification`

List one contact's delivered notifications, newest first.

```
ht-cli alerts list-contact-notification <id> [flags]
```

Returns a page of the notifications delivered to the contact in the path, with the same window and outcome filters, the same row shape and the same expansions as the account-wide list. Use it when the question is about one address; the collection read at /contact/notification serves cross-contact audits.

GET /contact/{id}/notification (listContactNotification)

Arguments:
  <id>	the contact id

Filters too long for a query string go in a body query: --query-file filter.json (POST /contact/{id}/notification/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| sentAt \| kind \| channel \| gateway \| contact \| monitor \| checkNumber \| ... 3 more) (repeatable) |
| `--from` | int64 | Window start, Unix seconds. |
| `--limit` | int64 | Rows to return. |
| `--outcome` | stringSlice | Which delivery outcomes to include. (billingFailed \| blocked \| cancelled \| grouped \| insufficientBalance \| noProfile \| renderFailed \| sendFailed \| ... 3 more) (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /contact/{id}/notification/q: inline JSON, @file, or - |
| `--to` | int64 | Window end, Unix seconds. |

## `ht-cli alerts list-monitor`

List the contacts this monitor alerts, with each contact's alert-type set.

```
ht-cli alerts list-monitor <monitor-id> [flags]
```

Returns who is alerted when this monitor changes state - one row per contact, each carrying the set of alert types (up/down/repeatedlyDown) that contact receives. Set the subscription with PUT /monitor/{monitorId}/alert/{contactId}. The contact's address is included only when the token also carries a contact read scope - a monitor-scoped token sees which contact it is, not how to reach it.

GET /monitor/{monitorId}/alert (listMonitorAlert)

Arguments:
  <monitor-id>	the monitor id

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/{monitorId}/alert/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (contact \| alertTypes \| created) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/{monitorId}/alert/q: inline JSON, @file, or - |

## `ht-cli alerts list-notification`

List notifications delivered to the account, newest first.

```
ht-cli alerts list-notification [flags]
```

Returns a page of delivered notifications, optionally narrowed by time window, contact or delivery outcome, so a client can audit what was actually sent and when. Every row names the monitor that caused it and the number of the check behind it, which is what makes local grouping possible, and carries the rendered subject with a short plain-text preview of the body. Read one notification by id when the whole body is wanted. Ask for expand=monitor.settings (or monitor.subscription / monitor.lastIncident / monitor.maintenance) to embed the monitor's own blocks in the row's monitor object.

GET /contact/notification (listNotification)

Filters too long for a query string go in a body query: --query-file filter.json (POST /contact/notification/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--contact` | stringSlice | Contact ids - the one relation the log is indexed by. (repeatable) |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Declaration only - the VALUE is read from the raw query (a bound string cannot tell ?expand= from an absent parameter, and keeps only the first of a... (monitor \| monitor.settings \| monitor.subscription \| monitor.lastIncident \| monitor.maintenance) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| sentAt \| kind \| channel \| gateway \| contact \| monitor \| checkNumber \| ... 3 more) (repeatable) |
| `--from` | int64 | Window start, Unix seconds. |
| `--limit` | int64 | Rows to return. |
| `--outcome` | stringSlice | Delivery outcomes, as the pipeline records them. (billingFailed \| blocked \| cancelled \| grouped \| insufficientBalance \| noProfile \| renderFailed \| sendFailed \| ... 3 more) (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /contact/notification/q: inline JSON, @file, or - |
| `--to` | int64 | Window end, Unix seconds. |

## `ht-cli alerts list-subscription`

List every alert subscription on the account, flat.

```
ht-cli alerts list-subscription [flags]
```

Returns every (monitor, contact) alert pair the account holds as one flat list - both sides identified on every row, with the pair's alert-type set. Narrow by the entity-prefixed filters of either side: monitor.id, monitor.type, monitor.tag, monitor.url (+monitor.like), monitor.q, and contact.id, contact.type, contact.confirmed, contact.q. The one read that audits the whole account's alert wiring; the nested per-monitor and per-contact lists answer one parent's slice.

GET /alert (listAlertSubscription)

Filters too long for a query string go in a body query: --query-file filter.json (POST /alert/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--contact-confirmed` | bool | See this operation's description for how this parameter narrows the result. |
| `--contact-id` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--contact-q` | string | See this operation's description for how this parameter narrows the result. |
| `--contact-type` | stringSlice | See this operation's description for how this parameter narrows the result. (email \| sms \| voiceCall \| http \| telegram \| viber \| facebook \| googleChat \| ... 3 more) (repeatable) |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| monitor \| contact \| alertTypes \| created) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--monitor-id` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--monitor-like` | bool | See this operation's description for how this parameter narrows the result. |
| `--monitor-q` | string | See this operation's description for how this parameter narrows the result. |
| `--monitor-tag` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--monitor-type` | stringSlice | See this operation's description for how this parameter narrows the result. (http \| waterfall \| ping \| port \| domainExp \| sslExp \| dnsbl \| webRisk \| ... 6 more) (repeatable) |
| `--monitor-url` | stringSlice | See this operation's description for how this parameter narrows the result. (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /alert/q: inline JSON, @file, or - |

## `ht-cli alerts list-type`

List the alert types and the alert delays a subscription may use.

```
ht-cli alerts list-type [flags]
```

Returns the fixed vocabulary of alert types a subscription can name, and the alert delay values the account may choose from. Use it to populate a picker or to validate a value before writing a subscription, rather than hard-coding a vocabulary that can grow.

GET /alert/type (listAlertType)

Filters too long for a query string go in a body query: --query-file filter.json (POST /alert/type/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /alert/type/q: inline JSON, @file, or - |

## `ht-cli alerts resend-notification`

Resend a scheduled report for one period.

```
ht-cli alerts resend-notification [flags]
```

Rebuilds the periodic report of one frequency for one period and emails it to the contacts subscribed to it, exactly as the schedule would have. Use it when a scheduled report did not arrive or was deleted. It is addressed by the schedule (frequency + any instant inside the period) rather than by a delivered notification, because the delivery log records that a report was sent and not which schedule produced it. The rebuild runs on the report service, so this answers immediately and the mail follows; the period must be within the last three years.

POST /contact/notification/resend (resendNotification)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli alerts set-contact`

Set the alert-type set for this contact-and-monitor pair.

```
ht-cli alerts set-contact <id> <monitor-id> [flags]
```

The contact-side mirror of setMonitorAlert: idempotently sets which alert types this contact receives for the monitor in the path. `alertTypes` is the EXACT desired set (at least one); use DELETE to remove the subscription. Answers with the resulting subscription.

PUT /contact/{id}/alert/{monitorId} (setContactAlert)

Arguments:
  <id>	the contact id
  <monitor-id>	the monitor id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli alerts set-monitor`

Set the alert-type set for this monitor-and-contact pair.

```
ht-cli alerts set-monitor <monitor-id> <contact-id> [flags]
```

Idempotently sets which alert types this contact receives for this monitor. `alertTypes` is the EXACT desired set (at least one); use DELETE to remove the subscription. Answers with the resulting subscription. The contact's address is included only when the token also carries a contact read scope - a monitor-scoped token sees which contact it is, not how to reach it.

PUT /monitor/{monitorId}/alert/{contactId} (setMonitorAlert)

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
