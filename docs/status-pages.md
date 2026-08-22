# `ht-cli status-pages`

Public status pages, their incidents, templates and subscribers

## `ht-cli status-pages add-incident-timeline-entry`

Append an entry to a declared incident's timeline.

```
ht-cli status-pages add-incident-timeline-entry <id> <incident-id> [flags]
```

Appends {state, message} to the append-only timeline - the way an incident progresses and the way it resolves (a resolved state stamps resolvedAt, first time only; appending to a resolved incident is allowed and reads as a follow-up note). Appending notifies the page's confirmed subscribers, scoped by the incident's own affected components. Answers the incident as it now stands. An Idempotency-Key is REQUIRED here: this call notifies the page's subscribers, so a retry without one announces the same thing twice.

POST /statuspage/{id}/incident/{incidentId}/timeline (addStatusPageIncidentTimelineEntry)

Arguments:
  <id>	the status page id
  <incident-id>	the incident id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required. A client-chosen key, unique per logical request. (required) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli status-pages add-subscriber`

Add a push channel to a status page.

```
ht-cli status-pages add-subscriber <id> [flags]
```

Registers a webhook, Slack or Teams channel for the page's incident updates, optionally scoped to one component. Adding an EMAIL subscriber is refused by design - an address joins only through the public page's own double-opt-in, never by owner fiat. The plan's confirmed-subscriber cap is enforced; the refusal names the numbers.

POST /statuspage/{id}/subscriber (addStatusPageSubscriber)

Arguments:
  <id>	the status page id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli status-pages create`

Create a status page.

```
ht-cli status-pages create [flags]
```

Creates a page at the given slug - its permanent public address - optionally with its initial settings and component set in the same call. The slug is lowercase letters, digits and single hyphens, unique across the product, and not changeable through this surface. How many pages a plan allows is enforced here; the refusal names the numbers.

POST /statuspage (createStatusPage)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli status-pages create-incident`

Declare an incident or a scheduled maintenance on a status page.

```
ht-cli status-pages create-incident <id> [flags]
```

Declares what is happening: a title, the initial state (which seeds the timeline with the message), the affected components (this page's own), and for kind=maintenance the scheduled window in Unix seconds. This is the CI/automation door for incident communication. Declaring notifies the page's confirmed subscribers (email + webhook/Slack/Teams channels, component-scoped) exactly as a dashboard declare does - best-effort, a delivery failure never fails the write. An Idempotency-Key is REQUIRED here: this call notifies the page's subscribers, so a retry without one announces the same thing twice.

POST /statuspage/{id}/incident (createStatusPageIncident)

Arguments:
  <id>	the status page id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required. A client-chosen key, unique per logical request. (required) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli status-pages create-template`

Save an incident template.

```
ht-cli status-pages create-template <id> [flags]
```

Saves a preset {title, message, defaultImpact?} for future declarations. Templates have no update by design - save a new one and delete the old.

POST /statuspage/{id}/template (createStatusPageTemplate)

Arguments:
  <id>	the status page id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli status-pages delete`

Delete a status page.

```
ht-cli status-pages delete <id> [flags]
```

Removes the page with its components, declared incidents and subscribers, and frees its slug. 404 for a page that does not exist or belongs to another account, so a 200 always means it existed and is gone. The receipt names the slug that is free again and counts what went with the page - components, declared incidents, subscribers and templates.

DELETE /statuspage/{id} (deleteStatusPage)

Arguments:
  <id>	the status page id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli status-pages delete-incident`

Delete a declared incident.

```
ht-cli status-pages delete-incident <id> <incident-id> [flags]
```

Removes the incident and its timeline from the page, with no subscriber notice - the same deliberate choice the dashboard makes. 404 for absent/foreign/wrong-page. The receipt counts the timeline entries that went with it.

DELETE /statuspage/{id}/incident/{incidentId} (deleteStatusPageIncident)

Arguments:
  <id>	the status page id
  <incident-id>	the incident id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli status-pages delete-subscriber`

Remove a subscriber from a status page.

```
ht-cli status-pages delete-subscriber <id> <subscriber-id> [flags]
```

Removes one subscriber of any kind. 404 when the id is not this page's, so a 200 always means something was removed.

DELETE /statuspage/{id}/subscriber/{subscriberId} (deleteStatusPageSubscriber)

Arguments:
  <id>	the status page id
  <subscriber-id>	the subscriber id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli status-pages delete-template`

Delete an incident template.

```
ht-cli status-pages delete-template <id> <template-id> [flags]
```

Removes one of THIS page's templates. 404 when the template is not this page's, so a 200 always means it existed here and is gone.

DELETE /statuspage/{id}/template/{templateId} (deleteStatusPageTemplate)

Arguments:
  <id>	the status page id
  <template-id>	the template id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli status-pages get`

Read one status page's configuration and components.

```
ht-cli status-pages get <id> [flags]
```

Returns the page's full configuration - the settings block (appearance, behaviour and the feature set, published as named tokens rather than a bitmask) and the ordered component set. A page that does not exist and one that belongs to another account answer the same 404. It is a superset of the list row: componentCount, unresolvedIncidents and created are carried here too, so a client never has to keep a list row to redraw a page it opened.

GET /statuspage/{id} (getStatusPage)

Arguments:
  <id>	the status page id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| slug \| title \| componentCount \| unresolvedIncidents \| hasPassword \| created \| settings \| ... 2 more) (repeatable) |

## `ht-cli status-pages get-incident`

Read one declared incident with its full timeline.

```
ht-cli status-pages get-incident <id> <incident-id> [flags]
```

Returns the incident - status, kind, impact, affected components, the append-only update timeline and the postmortem when one is written. 404 when it is absent, another account's, or another page's.

GET /statuspage/{id}/incident/{incidentId} (getStatusPageIncident)

Arguments:
  <id>	the status page id
  <incident-id>	the incident id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| title \| state \| kind \| impact \| created \| resolvedAt \| scheduledStart \| ... 5 more) (repeatable) |

## `ht-cli status-pages list`

List the account's status pages.

```
ht-cli status-pages list [flags]
```

Returns every status page the account owns - slug, title, component count, unresolved declared incidents and whether a password gates the public view. The public rendering itself lives at the page's own address; this surface manages the configuration. Sortable: sort=created|title|slug, optionally suffixed :asc or :desc.

GET /statuspage (listStatusPage)

Filters too long for a query string go in a body query: --query-file filter.json (POST /statuspage/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| slug \| title \| componentCount \| unresolvedIncidents \| hasPassword \| created) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /statuspage/q: inline JSON, @file, or - |
| `--sort` | string | created \| title \| slug, optionally suffixed :asc/:desc. (created \| title \| slug \| created:asc \| created:desc \| title:asc \| title:desc \| slug:asc \| ... 1 more) |

## `ht-cli status-pages list-incident`

List a status page's declared incidents.

```
ht-cli status-pages list-incident <id> [flags]
```

Returns the incidents and maintenances the OWNER declared on this page, newest first, each with its status timeline and affected components. Distinct from the monitor-derived episodes at /monitor/incident: these are the page's own communication.

GET /statuspage/{id}/incident (listStatusPageIncident)

Arguments:
  <id>	the status page id

Filters too long for a query string go in a body query: --query-file filter.json (POST /statuspage/{id}/incident/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| title \| state \| kind \| impact \| created \| resolvedAt \| scheduledStart \| ... 5 more) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /statuspage/{id}/incident/q: inline JSON, @file, or - |

## `ht-cli status-pages list-subscriber`

List a status page's subscribers.

```
ht-cli status-pages list-subscriber <id> [flags]
```

Returns every subscriber - email addresses (with their confirmation state; they join through the public page's double-opt-in) and push channels (webhook/Slack/Teams) - each optionally scoped to one component.

GET /statuspage/{id}/subscriber (listStatusPageSubscriber)

Arguments:
  <id>	the status page id

Filters too long for a query string go in a body query: --query-file filter.json (POST /statuspage/{id}/subscriber/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| kind \| email \| url \| componentId \| confirmedAt \| created) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /statuspage/{id}/subscriber/q: inline JSON, @file, or - |

## `ht-cli status-pages list-template`

List a status page's incident templates.

```
ht-cli status-pages list-template <id> [flags]
```

Returns the page's saved incident presets - title, message template and default impact - for declaring quickly during a real event.

GET /statuspage/{id}/template (listStatusPageTemplate)

Arguments:
  <id>	the status page id

Filters too long for a query string go in a body query: --query-file filter.json (POST /statuspage/{id}/template/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| title \| message \| defaultImpact \| created) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /statuspage/{id}/template/q: inline JSON, @file, or - |

## `ht-cli status-pages set-components`

Replace a status page's whole component set.

```
ht-cli status-pages set-components <id> [flags]
```

A snapshot write, not a diff: send every component to keep, in display order - each either monitored (its monitor must be the account's own) or third-party (own name, optional manual state). Carry each existing row's id so per-component subscriptions survive the save. How many components a plan allows is enforced; a page already over its cap may re-save and reorder, only growing past it is refused.

PUT /statuspage/{id}/component (setStatusPageComponents)

Arguments:
  <id>	the status page id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli status-pages update`

Change a status page's title and/or settings.

```
ht-cli status-pages update <id> [flags]
```

Patches the page: send title and/or any settings members - only what is sent changes, and an explicit null clears a clearable member. Features are named tokens; sending the array replaces the whole feature set. Answers the page as it now stands.

PATCH /statuspage/{id} (updateStatusPage)

Arguments:
  <id>	the status page id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli status-pages update-incident`

Fix a declared incident's title, kind or components; set its postmortem.

```
ht-cli status-pages update-incident <id> <incident-id> [flags]
```

Patches what the incident IS - title, kind, affected components - without touching its timeline, and sets/replaces/clears the postmortem write-up. Progress is NOT a patch: append a timeline update instead, and resolve by appending a resolved update.

PATCH /statuspage/{id}/incident/{incidentId} (updateStatusPageIncident)

Arguments:
  <id>	the status page id
  <incident-id>	the incident id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

---

[Back to the index](README.md)
