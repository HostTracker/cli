# `ht jobs`

Long-running batch jobs started by the bulk operations

## `ht jobs cancel`

Cancel a queued or running asynchronous operation.

```
ht jobs cancel <id> [flags]
```

Asks a job that has not finished to stop, and answers with its state and the results it produced before stopping. Work already completed is not undone, and the job itself is not removed - it stays readable for the rest of its retention window. Cancelling a job that is already cancelling is safe and changes nothing; a job that has already reached a final state cannot be cancelled and says so.

POST /job/{id}/cancel (cancelJob)

Arguments:
  <id>	the job id

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--limit` | int64 | Rows to return. |

## `ht jobs get`

Poll an asynchronous operation's state and per-item results.

```
ht jobs get <id> [flags]
```

Returns the state, progress and a page of per-item results for any asynchronous operation started anywhere on this API, whatever created it. Poll it until the state is terminal. Large result sets page inside the response, so keep following the cursor until none is returned. The scope required is the one the creating operation used, so a token that could start the job can always read it. While the job is still running the response carries a Retry-After header saying how long to wait before polling again; a finished job carries none.

GET /job/{id} (getJob)

Arguments:
  <id>	the job id

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| kind \| scope \| state \| progress \| summary \| cancelRequested \| created \| ... 10 more) (repeatable) |
| `--limit` | int64 | Rows to return. |

## `ht jobs list`

List this account's recent asynchronous operations.

```
ht jobs list [flags]
```

Returns a page of the asynchronous operations this account has started, newest first, optionally narrowed by kind or state. Reach for it when a job id was lost between starting the work and polling it, or to show what is currently running. Rows carry each job's state, progress and counts but not its per-item results - follow a row's results url to read those. Jobs are readable for seven days, which is what bounds this list.

GET /job (listJob)

Filters too long for a query string go in a body query: --query-file filter.json (POST /job/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (id \| kind \| scope \| state \| progress \| summary \| cancelRequested \| created \| ... 6 more) (repeatable) |
| `--kind` | stringSlice | Job kinds, e.g. monitor.bulkCreate - ANY-OF. (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /job/q: inline JSON, @file, or - |
| `--state` | stringSlice | Which lifecycle states to include. (queued \| running \| succeeded \| partial \| failed \| cancelled \| interrupted) (repeatable) |

## `ht jobs resume`

Continue an asynchronous operation that was interrupted.

```
ht jobs resume <id> [flags]
```

Restarts an operation whose state is interrupted - one that stopped because the server it was running on stopped, not because anything about the work failed. Everything it had already done stands, and the continuation skips those items, so resuming is safe and never repeats work. Only an interrupted job can be resumed: one that is still running needs nothing but another poll, and a few kinds cannot be continued at all because the request they were carrying is not stored - those say so and ask you to submit again.

POST /job/{id}/resume (resumeJob)

Arguments:
  <id>	the job id

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--idempotency-key` | string | Optional. A client-chosen key, unique per logical request. |
| `--limit` | int64 | Rows to return. |

## `ht jobs wait`

Poll a job until it reaches a terminal state

```
ht jobs wait <id> [flags]
```

Poll a job until it reaches a terminal state.

The bulk operations answer 202 with a job id; this follows that job at the
pace the API asks for and prints it once it stops moving.

Reaching a terminal state is a success whatever that state is: succeeded,
partial, failed and cancelled all exit 0, and the printed state is what
says which. A job that was interrupted stops the wait as well, because it
makes no further progress until ht jobs resume continues it.

| Flag | Type | Description |
|---|---|---|
| `--follow-interrupts` | bool | keep polling an interrupted job instead of returning it |
| `--interval` | duration | poll interval when the answer asks for none (default 2s) |
| `--wait-timeout` | duration | how long to keep polling |

---

[Back to the index](README.md)
