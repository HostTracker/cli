# `ht monitoring-locations`

The monitoring locations checks can run from

## `ht monitoring-locations list`

List the monitoring locations checks can run from.

```
ht monitoring-locations list [flags]
```

Returns a page of monitoring locations with each one's country, provider, capabilities and pool memberships. Filter by country, pool or capability to find the locations that suit a particular check before configuring where a monitor runs from. The fleet is shared infrastructure, so this list is the same for every account.

GET /agent (listAgent)

Filters too long for a query string go in a body query: --query-file filter.json (POST /agent/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--capability` | stringSlice | icmp \| browser \| internal. (browser \| icmp \| internal) (repeatable) |
| `--country` | stringSlice | Datacenter country names. (repeatable) |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| name \| city \| region \| country \| lat \| lon \| provider \| ... 7 more) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--pool` | stringSlice | Pool ids - a location is kept when it belongs to any of them (parents populated). (repeatable) |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /agent/q: inline JSON, @file, or - |

## `ht monitoring-locations list-ip`

List the IP addresses monitoring checks originate from.

```
ht monitoring-locations list-ip [flags]
```

Returns the addresses checks are sent from, for allow-listing in a firewall or in a monitored service's access rules. It needs no token - an allow-listing script frequently runs before any credential exists on the machine it is provisioning - and a token is honoured if you send one: the answer is identical, and the call is metered on your account's own reference bucket instead of the one shared by everyone calling from your address. Every response carries the rate-limit headers.

GET /agent/ip (listAgentIp)

Filters too long for a query string go in a body query: --query-file filter.json (POST /agent/ip/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--country` | stringSlice | Datacenter country names. (repeatable) |
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--family` | stringSlice | Restrict IP families - ANY-OF - ipv4 or ipv6. (ipv4 \| ipv6) (repeatable) |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (ip \| family \| country \| addedAt) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /agent/ip/q: inline JSON, @file, or - |

## `ht monitoring-locations list-pool`

List the location pools, with counts, defaults and presets.

```
ht monitoring-locations list-pool [flags]
```

Returns the pools monitoring locations are grouped into, how many locations each offers per capability, and the account's own minimum-agent requirement, default selection and saved presets. Read it before choosing a monitor's locations: it is what tells you whether a selection can meet the minimum before a create refuses it for not doing so.

GET /agent/pool (listAgentPool)

Filters too long for a query string go in a body query: --query-file filter.json (POST /agent/pool/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| name \| agents \| children \| parents \| hidden \| priority \| agentIds) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /agent/pool/q: inline JSON, @file, or - |

---

[Back to the index](README.md)
