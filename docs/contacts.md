# `ht contacts`

Notification contacts, contact groups and confirmations

## `ht contacts bulk-delete`

Delete every contact a filter selects, as an asynchronous job.

```
ht contacts bulk-delete [flags]
```

Resolves the filter to a concrete set of contacts, answers with a job id and the number of contacts accepted, and deletes them one at a time - each in its own transaction, each with its own receipt of what was cascaded away. The set is fixed at submission: a contact that starts matching the filter afterwards is not touched. Check what a filter selects first with bulk-delete-validate; this call then acts on it, with no token to echo back. A filter that narrows by nothing is refused rather than taken to mean the whole account, and a selection larger than one operation carries is refused rather than trimmed. An Idempotency-Key is mandatory: this call answers 202 and then works asynchronously, so a retry after a timeout would otherwise start a second job over the same targets. The 202 carries a Retry-After header saying how long to wait before the first poll, sized on how much work was accepted.

POST /contact/bulk-delete (bulkDeleteContact)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required. A client-chosen key, unique per logical request. (required) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts bulk-delete-validate`

Check which contacts a delete filter selects, without deleting anything.

```
ht contacts bulk-delete-validate [flags]
```

Resolves a filter and answers how many contacts it selects together with a sample of them, identified by name and address. Nothing is written. It is the verification step of a bulk delete: run it, show a human what is about to go, then send the same filter to the delete itself, which acts on it directly. The answer is a snapshot - contacts created or removed between the two calls change what the delete resolves, which is why the delete reports its own count as well.

POST /contact/bulk-delete-validate (bulkDeleteValidateContact)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | replay key for this write (default: a fresh one per call) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts bulk-validate`

Check what a contact batch would do, without writing anything.

```
ht contacts bulk-validate [flags]
```

Takes the same body as the bulk write and answers a verdict per item instead of applying it: whether the item is valid, the errors it would be refused with, and whether the account's package has room for it - it fits, it would be created disabled, or it would fail. Nothing is written and no Idempotency-Key is needed. Use it to pre-flight an import before spending a key on it, since the write applies items one at a time and can partially succeed.

POST /contact/bulk-validate (bulkValidateContact)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | replay key for this write (default: a fresh one per call) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts bulk-write`

Create, update and delete many contacts as one asynchronous job.

```
ht contacts bulk-write [flags]
```

Submits a batch of contact creations, updates and deletions as a single job and answers immediately with a job id to poll for per-item results. Items are applied one at a time, each in its own transaction, so a batch can partially succeed and the job reports exactly which items landed. Check a batch in advance with bulk-validate. A create the account's package has no room for is created paused rather than refused, unless onOverlimit says otherwise. An Idempotency-Key is mandatory, because a retry would duplicate rows and re-send paid confirmation messages. The 202 carries a Retry-After header saying how long to wait before the first poll, sized on how much work was accepted.

POST /contact/bulk (bulkWriteContact)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required. A client-chosen key, unique per logical request. (required) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts create`

Create a contact, or bind to the matching one that already exists.

```
ht contacts create [flags]
```

Creates a delivery contact and, for a channel that needs confirming, issues and sends its confirmation code in the same call. When a contact with the same type, address, gateway and alert delay already exists, the request binds to it instead of creating a duplicate - and the status says which happened, 201 for a new row and 200 for a bind. Alert and report subscriptions can be wired in the same request, riding this same contact:write.

POST /contact (createContact)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Required for some requests, optional for the rest - see this operation's description for the condition. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts create-group`

Create a contact group.

```
ht contacts create-group [flags]
```

Creates a named preset of contacts and their event selections. The name is unique per account (case-insensitively); at least one contact with at least one event is required; every contact must be the account's own. Events span both subscription kinds in one vocabulary - the three alert types and the five report frequencies.

POST /contact/group (createContactGroup)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts delete`

Delete a contact and receive a receipt of what was removed.

```
ht contacts delete <id> [flags]
```

Deletes a contact along with every alert and report subscription pointing at it, and answers with a receipt counting what was cascaded away. Use it to retire a delivery channel; the counts exist nowhere else once the delete has run, which is why the response carries a body rather than nothing.

DELETE /contact/{id} (deleteContact)

Arguments:
  <id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht contacts delete-group`

Delete a contact group.

```
ht contacts delete-group <id> [flags]
```

Removes the group and its membership rows. Deleting a group never touches any live subscription - membership was a preset, not a subscription. 404 for a group that does not exist or belongs to another account, so a 200 always means it existed and is gone. The receipt names the group and counts the contacts its preset listed; the contacts themselves are untouched.

DELETE /contact/group/{id} (deleteContactGroup)

Arguments:
  <id>	the group id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht contacts get`

Retrieve one contact by id.

```
ht contacts get <id> [flags]
```

Returns one contact in exactly the row shape the list endpoint uses, including its delivery address, confirmation state, alert delay and - for a channel that is charged per message - the price of one send. Use it when the id is already known rather than filtering the list down to it. A contact that does not exist and one belonging to another account answer the same not-found.

GET /contact/{id} (getContact)

Arguments:
  <id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (subscription \| template \| group \| summary \| count) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| type \| name \| address \| confirmed \| overlimited \| alertDelay \| sendCost \| ... 14 more) (repeatable) |

## `ht contacts get-group`

Read one contact group.

```
ht contacts get-group <id> [flags]
```

Returns a single contact group with its full membership snapshot, in exactly the list's row shape. A group that does not exist and one that belongs to another account answer the same 404.

GET /contact/group/{id} (getContactGroup)

Arguments:
  <id>	the group id

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| name \| created \| items) (repeatable) |

## `ht contacts list`

List the account's contacts, filtered and cursor-paginated.

```
ht contacts list [flags]
```

Returns a page of contacts - the addresses alerts and reports are delivered to - narrowed by id, type, confirmation state or a free-text search over name and address. Use expand to embed each contact's subscriptions, its message template or the groups it belongs to, and updatedSince for an incremental poll. Fetch one contact by id when only one is wanted. Order it with sort=created|name|address, optionally suffixed :asc or :desc (sort=name:desc); without a suffix created reads newest-first and the text columns A to Z. There is no separate order parameter.

GET /contact (listContact)

Filters too long for a query string go in a body query: --query-file filter.json (POST /contact/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--confirmed` | bool | Restrict to confirmed, or to unconfirmed. |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (subscription \| template \| group \| summary \| count) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| type \| name \| address \| confirmed \| overlimited \| alertDelay \| sendCost \| ... 14 more) (repeatable) |
| `--id` | stringSlice | Explicit ids. A PRESENT-but-empty value selects NOTHING; an absent parameter is no filter. (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--q` | string | Case-insensitive substring over name AND address. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /contact/q: inline JSON, @file, or - |
| `--sort` | string | created \| name \| address, optionally suffixed with :asc or :desc (sort=name:desc). (created \| name \| address \| created:asc \| created:desc \| name:asc \| name:desc \| address:asc \| ... 1 more) |
| `--type` | stringSlice | One or more contact-type tokens; a row matching any of them is returned. (email \| sms \| voiceCall \| http \| telegram \| viber \| facebook \| googleChat \| ... 3 more) (repeatable) |
| `--updated-since` | string | Return only what changed since this point - either Unix seconds, or the syncCursor a previous response returned. |

## `ht contacts list-group`

List the account's contact groups.

```
ht contacts list-group [flags]
```

Returns every contact group with its full membership snapshot - each member as the contact's identifying projection plus its event selection. A group is a PASSIVE preset: applying one means reading it and writing the subscriptions you want through the subscription endpoints; membership by itself subscribes nothing. Sortable: sort=name|created, optionally suffixed :asc or :desc (default name ascending, the list's shipped order).

GET /contact/group (listContactGroup)

Filters too long for a query string go in a body query: --query-file filter.json (POST /contact/group/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| name \| created \| items) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /contact/group/q: inline JSON, @file, or - |
| `--sort` | string | name \| created, optionally suffixed :asc/:desc. (name \| created \| name:asc \| name:desc \| created:asc \| created:desc) |

## `ht contacts list-type`

List the contact types, their gateways and their capabilities.

```
ht contacts list-type [flags]
```

Returns the catalogue of contact types: which can be created directly, which need confirmation, which can receive scheduled reports, the gateways each offers, and the alert delays the account may choose from. Use it to build a contact form or to validate a type before a create - the creatable set is not fixed forever, so reading it beats hard-coding it.

GET /contact/type (listContactType)

Filters too long for a query string go in a body query: --query-file filter.json (POST /contact/type/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /contact/type/q: inline JSON, @file, or - |

## `ht contacts resend-confirmation`

Resend the confirmation code for an unconfirmed contact.

```
ht contacts resend-confirmation <id> [flags]
```

Reissues the confirmation code for a contact, for when the code sent at creation never arrived or its validity window has passed. A code that is still valid is sent again unchanged rather than replaced, so a resend can never invalidate a code the recipient is already reading. Resends are rate-limited per contact and a refusal names when to try again.

POST /contact/{id}/confirmation (resendContactConfirmation)

Arguments:
  <id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |

## `ht contacts set-groups`

Set the contact groups one contact belongs to.

```
ht contacts set-groups <id> [flags]
```

Replaces a contact's whole group membership in one call: the body names every group the contact should be in, so a group it omits is left and an empty list removes the contact from all of them. Reach for it instead of patching each group in turn - that costs one write per group and rewrites memberships of contacts the caller was not changing. Each entry may name the events to join with; omit them to join with the events the group already carries. A group is a passive preset either way, so joining one subscribes the contact to nothing by itself.

PUT /contact/{id}/group (setContactGroups)

Arguments:
  <id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts test`

Send a real test alert to a confirmed contact.

```
ht contacts test <id> [flags]
```

Sends a test notification through the same delivery pipeline production alerts use, so it proves actual delivery rather than only formatting, and reports the outcome synchronously rather than handing back a job. Use it while setting an integration up. Test sends are rate-limited, and a contact that is not yet confirmed or is suppressed by the account's plan is refused.

POST /contact/{id}/test (testContact)

Arguments:
  <id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts update`

Partially update a contact and get the updated contact back.

```
ht contacts update <id> [flags]
```

Updates one or more members of a contact; a member the body omits stays as it was, and an explicit null clears an optional one. Changing the delivery address resets confirmation and sends a fresh code, because the new address has never proved it belongs to you. A NEW email address is also checked for deliverability, exactly as on create: a domain with no mail host is refused as `validation_failed` with reason `undeliverable_domain` (fail-open on DNS trouble; a body that does not touch `address` is never checked). A contact's type cannot change after creation - create another contact instead.

PATCH /contact/{id} (updateContact)

Arguments:
  <id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts update-group`

Change a contact group's name and/or replace its membership.

```
ht contacts update-group <id> [flags]
```

Patches a group: send `name` to rename, `items` to REPLACE the whole membership snapshot, or both. A group is a snapshot, so membership has no per-row diff - a present items member is the exact desired set. Answers with the group as it now stands.

PATCH /contact/group/{id} (updateContactGroup)

Arguments:
  <id>	the group id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

## `ht contacts verify`

Confirm a contact with the code it received.

```
ht contacts verify <id> [flags]
```

Submits the code delivered to a contact and, when it is correct and still valid, marks the contact confirmed and answers with the contact as it now stands. A wrong or expired code is refused with how many attempts remain and when the code expires, so a client can tell 'try again' from 'ask for a new one'. Confirming an already-confirmed contact is reported distinctly from a fresh success.

POST /contact/{id}/confirmation/verify (verifyContact)

Arguments:
  <id>	the contact id

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

---

[Back to the index](README.md)
