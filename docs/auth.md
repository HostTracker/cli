# `ht auth`

Store, inspect and remove the API token

Store, inspect and remove the API token.

The token is kept in the profile file (ht config path), which is written
0600. A token given with --token, or found in HT_TOKEN, wins over the
stored one for that command, so a CI job needs no file at all.

## `ht auth login`

Store an API token in a profile

```
ht auth login [flags]
```

Store an API token in a profile.

The token is read from --token, from HT_TOKEN, from standard input with
--token-stdin, or from a prompt that does not echo. It is verified against
GET /account before it is written, unless --verify=false.

  ht auth login                              prompt for a token
  ht auth login --token-stdin < token.txt    read it from a file
  ht auth login --profile staging            store it under another profile

| Flag | Type | Description |
|---|---|---|
| `--token-stdin` | bool | read the token from standard input |
| `--verify` | bool | check the token against the API before storing it |

## `ht auth logout`

Remove a profile's stored token

```
ht auth logout
```

## `ht auth status`

Report which credential is in force and whether it works

```
ht auth status [flags]
```

| Flag | Type | Description |
|---|---|---|
| `--offline` | bool | report the stored settings without calling the API |

---

[Back to the index](README.md)
