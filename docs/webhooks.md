# `ht-cli webhooks`

Webhook endpoints, their deliveries and test sends

## `ht-cli webhooks create`

Register a webhook to receive signed event deliveries.

```
ht-cli webhooks create [flags]
```

Registers an endpoint that receives signed HTTP deliveries for the events and monitors you name. This response is the only place the full signing secret is ever returned - generate one here or supply your own, and store it, because every later read shows only that a secret is set. A url that already has a webhook is refused rather than silently duplicated. Optional custom request headers travel with every delivery; names in the HT- and webhook- namespaces are reserved for the delivery's own headers and are refused.

POST /webhook (createWebhook)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli webhooks delete`

Unregister a webhook and stop its deliveries.

```
ht-cli webhooks delete <id> [flags]
```

Deletes a webhook and answers with a receipt naming how many monitors it was addressed to and how many deliveries were still waiting to retry and are now dropped. Use it to retire an integration for good; to pause one temporarily, set its enabled flag to false instead.

DELETE /webhook/{id} (deleteWebhook)

Arguments:
  <id>	the webhook id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli webhooks get`

Get one registered webhook by id.

```
ht-cli webhooks get <id> [flags]
```

Returns a single webhook in the shape the list endpoint uses, minus the signing secret's value. Use it to refresh one webhook's state - its enabled flag, its failure count, when it last delivered - without re-reading the whole account's list.

GET /webhook/{id} (getWebhook)

Arguments:
  <id>	the webhook id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| url \| events \| scope \| name \| enabled \| disabledReason \| consecutiveFailures \| ... 5 more) (repeatable) |

## `ht-cli webhooks list`

List the account's registered webhooks.

```
ht-cli webhooks list [flags]
```

Returns every webhook registered on the account with its enabled state, event subscriptions and delivery health - never the signing secret's value, only whether one is set. The same response publishes the full catalogue of event types a webhook may subscribe to, so a management UI needs one call rather than two. Sortable (sort=created|updated|name|url, optionally suffixed :asc or :desc), and sync-shaped: updatedSince= (Unix seconds or a previous response's syncCursor) narrows to rows edited since - `updated` is edit-faithful on this surface.

GET /webhook (listWebhook)

Filters too long for a query string go in a body query: --query-file filter.json (POST /webhook/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| url \| events \| scope \| name \| enabled \| disabledReason \| consecutiveFailures \| ... 5 more) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /webhook/q: inline JSON, @file, or - |
| `--sort` | string | created \| updated \| name \| url, optionally suffixed :asc/:desc. (created \| updated \| name \| url \| created:asc \| created:desc \| updated:asc \| updated:desc \| ... 4 more) |
| `--updated-since` | string | Return only what changed since this point - either Unix seconds, or the syncCursor a previous response returned. |

## `ht-cli webhooks list-delivery`

List recent deliveries for one webhook.

```
ht-cli webhooks list-delivery <id> [flags]
```

Returns a page of recent deliveries for one webhook, each with its outcome, every attempt it took, the status code the endpoint answered and a short excerpt of its response - which is usually enough to see why a delivery failed. Filter by time window, event or outcome. A delivery that is still retrying reads as pending and carries the time of its next attempt. Deliveries are kept for a bounded window; the webhook's own record of its last delivery and failure count is always available.

GET /webhook/{id}/delivery (listWebhookDelivery)

Arguments:
  <id>	the webhook id

Filters too long for a query string go in a body query: --query-file filter.json (POST /webhook/{id}/delivery/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--event` | stringSlice | Event tokens. (monitor.down \| monitor.up \| monitor.repeatedlyDown \| incident.opened \| incident.closed \| monitor.created \| monitor.updated \| monitor.deleted \| ... 7 more) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| event \| occurredAt \| attempts \| outcome \| nextRetryAt \| payloadDigest \| resourceId) (repeatable) |
| `--from` | int64 | Window start, Unix seconds. |
| `--limit` | int64 | Rows to return. |
| `--outcome` | stringSlice | Delivery outcomes. (pending \| delivered \| failed \| dropped) (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /webhook/{id}/delivery/q: inline JSON, @file, or - |
| `--to` | int64 | Window end, Unix seconds. |

## `ht-cli webhooks redeliver`

Resend a previously recorded webhook delivery.

```
ht-cli webhooks redeliver <id> <delivery-id> [flags]
```

Resends the exact payload of a recorded delivery with a freshly signed timestamp and reports the outcome synchronously. Use it to recover a delivery an endpoint missed while it was down, once the endpoint is healthy again. The delivery identifier is reused unchanged, so a consumer that deduplicates on it will not process the event twice.

POST /webhook/{id}/delivery/{deliveryId}/redeliver (redeliverWebhookDelivery)

Arguments:
  <id>	the webhook id
  <delivery-id>	the delivery id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht-cli webhooks test`

Send a synthetic test delivery to a webhook.

```
ht-cli webhooks test <id> [flags]
```

Sends one structurally realistic event to the webhook's endpoint, signed exactly the way a real delivery is, and reports the outcome synchronously - including the signature header value that was sent, so a receiving implementation can be debugged against it. A test is never retried, so a failure is visible immediately instead of being queued.

POST /webhook/{id}/test (testWebhook)

Arguments:
  <id>	the webhook id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli webhooks update`

Change a webhook's url, events, scope, name, headers or enabled state.

```
ht-cli webhooks update <id> [flags]
```

Updates the members the body names and leaves the rest alone. Re-enabling a webhook that auto-disabled also clears its failure count in the same call. Rotating the signing secret mints a new one, returns it once, and keeps the previous secret accepted for a grace window, so verification on the receiving side does not break the instant the rotation lands.

PATCH /webhook/{id} (updateWebhook)

Arguments:
  <id>	the webhook id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht-cli webhooks verify`

Check a webhook signature against the delivered body

```
ht-cli webhooks verify [flags]
```

Check a webhook signature against the delivered body.

The body is read from standard input (or --body-file), the headers from
--header or --headers-file, which takes a captured header block:

  HT-Signature: t=1735689600,v1=6f1c...
  HT-Delivery-Id: 018f...

The signature is computed over the RAW body, so it must be the exact bytes
the endpoint received: a re-serialised copy verifies against nothing.

  ht-cli webhooks verify --secret whsec_... --headers-file headers.txt < body.json

An accepted signature exits 0 and prints the event; a rejected one exits 5.
Both signature schemes the API sends are accepted, and several --secret
values may be passed at once, which is what a secret rotation needs.

| Flag | Type | Description |
|---|---|---|
| `--body-file` | string | file holding the raw delivered body (default: standard input) |
| `--header` | stringArray | one delivery header, as 'Name: value' (repeatable) |
| `--headers-file` | string | file holding the delivery's header block |
| `--secret` | stringArray | endpoint signing secret, whsec_ prefix included (repeatable during a rotation) |
| `--tolerance` | duration | how old a signature may be (default 5m) |

---

[Back to the index](README.md)
