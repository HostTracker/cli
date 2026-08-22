# `ht config`

Read and write the stored settings

Read and write the stored settings.

Each profile holds a token, a base-url and a default output format. The
file is written 0600 under the OS configuration directory; HT_CONFIG_DIR
moves it somewhere else, which is what a throwaway or CI profile wants.

## `ht config get`

Print one setting of the current profile

```
ht config get <key>
```

## `ht config list`

List the profiles and their settings

```
ht config list
```

## `ht config path`

Print where the settings are stored

```
ht config path
```

## `ht config set`

Write one setting of the current profile

```
ht config set <key> <value>
```

Write one setting of the current profile.

  ht config set base-url https://api2.host-tracker.com
  ht config set output json
  ht config set token "$HT_TOKEN"

Known keys: token, base-url, output.

---

[Back to the index](README.md)
