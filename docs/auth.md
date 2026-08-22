# `ht-cli auth`

Store, inspect and remove the API token

Store, inspect and remove the API token.

The token is kept in the profile file (ht-cli config path), which is written
0600. A token given with --token, or found in HT_TOKEN, wins over the
stored one for that command, so a CI job needs no file at all.

## `ht-cli auth login`

Store an API token in a profile

```
ht-cli auth login [flags]
```

Store an API token in a profile.

The token is read from --token, from HT_TOKEN, from standard input with
--token-stdin, or from a prompt that does not echo. It is verified against
GET /account before it is written, unless --verify=false.

  ht-cli auth login                              prompt for a token
  ht-cli auth login --token-stdin < token.txt    read it from a file
  ht-cli auth login --profile staging            store it under another profile

| Flag | Type | Description |
|---|---|---|
| `--token-stdin` | bool | read the token from standard input |
| `--verify` | bool | check the token against the API before storing it |

## `ht-cli auth logout`

Remove a profile's stored token

```
ht-cli auth logout
```

## `ht-cli auth status`

Report which credential is in force and whether it works

```
ht-cli auth status [flags]
```

| Flag | Type | Description |
|---|---|---|
| `--offline` | bool | report the stored settings without calling the API |

---

[Back to the index](README.md)
