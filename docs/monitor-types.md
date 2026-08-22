# `ht-cli monitor-types`

The catalogue of check types and their settings schema

## `ht-cli monitor-types get`

Get one monitor type's catalogue row and its full settings schema.

```
ht-cli monitor-types get <type> [flags]
```

Returns the complete JSON Schema describing one type's settings object, together with the catalogue row a picker already has, and - for the types that can also run attached to a parent monitor - the shape they take there. Use it to validate or generate a settings body for one type; the composite schema endpoint answers the same question for all types at once.

GET /monitor/type/{type} (getMonitorType)

Arguments:
  <type>	The kind of thing this is, from the relevant catalogue's closed vocabulary.

| Flag | Type | Description |
|---|---|---|
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (type \| schema \| attachedSchema) (repeatable) |

## `ht-cli monitor-types get-settings-schema`

Get one combined schema covering every monitor type's settings.

```
ht-cli monitor-types get-settings-schema
```

Returns a single JSON Schema document whose branches are every type's settings shape, selected by the monitor's type property, with each referenced shape defined once in a shared namespace. Use it when generating client types or a single schema artefact; fetch one type's schema instead when a form only ever edits one kind of monitor.

GET /monitor/type/schema (getMonitorSettingsSchema)

## `ht-cli monitor-types list`

List every monitor type, with its label, entitlement and constraints.

```
ht-cli monitor-types list [flags]
```

Returns the closed catalogue of monitor types: the display label, whether the account may create one, the minimum check interval, whether it needs a monitoring-location pool, and the presets it offers. Use it to drive a type picker or to validate a type before a create. Every listed type is creatable and every one has a full settings schema, so the catalogue never advertises something a write would then refuse.

GET /monitor/type (listMonitorType)

Filters too long for a query string go in a body query: --query-file filter.json (POST /monitor/type/q).

One page is returned by default. --all walks every page.

| Flag | Type | Description |
|---|---|---|
| `--cursor` | string | Opaque cursor from a previous response's nextCursor. |
| `--fields` | stringSlice | Which top-level members to keep on each row - fields=id,name. (type \| label \| creatable \| attachable \| minInterval \| fixedInterval \| requiresPool \| entitlement \| ... 3 more) (repeatable) |
| `--limit` | int64 | Rows to return. |
| `--query-file` | string | filter with a body query read from a file |
| `--query` | string | filter with a body query against /monitor/type/q: inline JSON, @file, or - |

---

[Back to the index](README.md)
