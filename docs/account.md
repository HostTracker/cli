# `ht-cli account`

The account, its quota and its usage

## `ht-cli account get`

Read the account: identity, package, usage, limits and status.

```
ht-cli account get [flags]
```

Returns the caller's account in one document - who it is, which package it holds, what it is using, the bounds every other endpoint enforces, and whether monitoring is currently running. Make it the first call an integration ever makes: the limits block is what lets a client size its later requests instead of discovering the bounds by being refused. Expand quota to fold the quota document in and save a second call.

GET /account (getAccount)

| Flag | Type | Description |
|---|---|---|
| `--expand` | stringSlice | Comma-separated names of the extra blocks to embed. (quota) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| login \| timezone \| language \| profile \| defaultAgentPools \| flags \| package \| ... 5 more) (repeatable) |

## `ht-cli account get-quota`

Read the account's current API quota headroom.

```
ht-cli account get-quota [flags]
```

Returns how much of each metered quota pool remains and when each window resets. Call it before starting a large batch to confirm there is room, rather than discovering the ceiling partway through. The same numbers ride every response as headers, so a client that reads those does not need this endpoint at all.

GET /account/quota (getAccountQuota)

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (limit \| used \| remaining \| resetAt \| pools \| tokenCap \| scopes \| apiEnabled) (repeatable) |

## `ht-cli account get-usage`

Read the account's current resource consumption per domain.

```
ht-cli account get-usage [flags]
```

Returns how much of each resource the account is using - under the singular keys `monitor`, `contact`, `report` (subscriptions) and `maintenance` (windows) - each as `{used, allowed}`, where `allowed: null` means no limit and `0` means a real cap of none. Call it to check headroom before a bulk create. This is the same `usage` block that rides the full `GET /account` object, exposed on its own so a client need not pull the whole account to poll consumption. Distinct from `quota`, which publishes the caps/entitlements you need BEFORE writing.

GET /account/usage (getAccountUsage)

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (monitor \| contact \| report \| maintenance) (repeatable) |

## `ht-cli account list-member`

List the contacts holding shared access to the account.

```
ht-cli account list-member [flags]
```

Returns a page of the people granted shared access to the account, with the rights each holds and whether the invitation has been accepted. Use it to audit or display who can act on the account's behalf. Access grants themselves are managed elsewhere; this surface is read-only.

GET /account/member (listAccountMember)

Filters too long for a query string go in a body query: --query-file filter.json (POST /account/member/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| userId \| contact \| rights \| state) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /account/member/q: inline JSON, @file, or - |

## `ht-cli account update`

Change the account's own details.

```
ht-cli account update [flags]
```

Updates the members the body names and leaves the rest alone: the account holder's name, company, phone and country, the time zone every schedule on the account is anchored to, the language notifications are written in, and the monitoring locations a new monitor starts with. Send an empty string to clear any of the four text fields, and defaultAgentPools as null to clear every default location. Answers the whole updated account, so nothing has to be re-read to confirm the change.

PATCH /account (updateAccount)

| Flag | Type | Description |
|---|---|---|
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--set` | stringArray | set one body member: --set name=api --set settings.interval=5 (repeatable) |

---

[Back to the index](README.md)
