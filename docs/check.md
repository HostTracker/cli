# `ht-cli check`

Run a one-off check without creating a monitor

Run a one-off check without creating a monitor.

The full instant-check surface, including the catalogue of types and the
history of past runs, is under ht-cli instant-checks.

## `ht-cli check run`

Start an instant check, and follow it to its result

```
ht-cli check run <url> [flags]
```

Start an instant check, and follow it to its result.

  ht-cli check run https://www.host-tracker.com --wait
  ht-cli check run example.com --type ping --country de --country us --wait
  ht-cli check run example.com:443 --type port --pool premium

Without --wait the command prints the 202 receipt, whose id is what
ht-cli instant-checks get <db-id> <id> reads later. With --wait it polls
at the pace the API asks for and prints the finished result.

--json supplies the whole request body for a shape the flags do not
reach, and the flags then fill in what it left out.

| Flag | Type | Description |
|---|---|---|
| `--city` | stringArray | run only from locations in this city (repeatable) |
| `--country` | stringArray | run only from locations in this country (repeatable) |
| `--idempotency-key` | string | replay key for this run (default: a fresh one per call) |
| `--json` | string | request body: inline JSON, @file, or - for standard input |
| `--pool` | stringArray | monitoring-location pool to run from (repeatable) |
| `--type` | string | which kind of check to run (default http) |
| `--wait-timeout` | duration | how long --wait keeps polling |
| `--wait` | bool | follow the check until it finishes and print the result |

---

[Back to the index](README.md)
