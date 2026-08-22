# `ht-cli config`

Read and write the stored settings

Read and write the stored settings.

Each profile holds a token, a base-url and a default output format. The
file is written 0600 under the OS configuration directory; HT_CONFIG_DIR
moves it somewhere else, which is what a throwaway or CI profile wants.

## `ht-cli config get`

Print one setting of the current profile

```
ht-cli config get <key>
```

## `ht-cli config list`

List the profiles and their settings

```
ht-cli config list
```

## `ht-cli config path`

Print where the settings are stored

```
ht-cli config path
```

## `ht-cli config set`

Write one setting of the current profile

```
ht-cli config set <key> <value>
```

Write one setting of the current profile.

  ht-cli config set base-url https://api2.host-tracker.com
  ht-cli config set output json
  ht-cli config set token "$HT_TOKEN"

Known keys: token, base-url, output.

---

[Back to the index](README.md)
