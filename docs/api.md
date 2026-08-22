# `ht api`

Send one request to any API address

Send one request to any API address.

The escape hatch for an address the generated commands do not model, and
the shortest way to reproduce a call from the reference documentation. The
request still rides the SDK, so the token, the idempotency key, the retry
ladder and the error mapping are the same as every other command's.

  ht api GET /account
  ht api GET /monitor --query limit=5 --query state=down
  ht api POST /monitor --set name=api --set type=http --set url=https://example.com
  ht api PATCH /monitor/<id> --json '{"enabled":false}'

## `ht api`

Send one request to any API address

```
ht api <method> <path> [flags]
```

Send one request to any API address.

The escape hatch for an address the generated commands do not model, and
the shortest way to reproduce a call from the reference documentation. The
request still rides the SDK, so the token, the idempotency key, the retry
ladder and the error mapping are the same as every other command's.

  ht api GET /account
  ht api GET /monitor --query limit=5 --query state=down
  ht api POST /monitor --set name=api --set type=http --set url=https://example.com
  ht api PATCH /monitor/<id> --json '{"enabled":false}'

| Flag | Type | Description |
|---|---|---|
| `--header` | stringArray | one request header, as 'Name: value' (repeatable) |
| `--idempotency-key` | string | replay key for this write (default: a fresh one per call) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--query` | stringArray | one query-string entry, as name=value (repeatable) |
| `--set` | stringArray | set one body member: --set name=api (repeatable) |

---

[Back to the index](README.md)
