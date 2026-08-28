# Logging

goss logs through [`log/slog`](https://pkg.go.dev/log/slog). The command line
tool builds a handler and writes records to stderr; as a library, goss writes
nothing until you give it a logger.

## Levels

goss uses the four standard `slog` levels and adds one below them.

| Level | Value | What goes here |
|-------|-------|----------------|
| `ERROR` | 8 | Critical errors that halt goss or significantly affect its functionality, requiring immediate intervention. |
| `WARN` | 4 | Non-critical issues that may require attention, such as overwritten keys or deprecated features. |
| `INFO` | 0 | General operational messages, useful for tasks where a more structured output is needed (e.g. `goss serve`). |
| `DEBUG` | -4 | Information useful for the goss user to debug. |
| `TRACE` | -8 | Detailed internal system activities useful for goss developers to debug. |

`TRACE` is `util.LevelTrace`. It is not a `slog` level, so a stock handler has no
name for it and renders it as `DEBUG-4`; see
[rendering TRACE](#rendering-trace-on-your-own-handler).

Selecting a level includes everything above it: `INFO` includes `WARN` and
`ERROR`. On the command line the level is set with
[`--loglevel`](cli.md#global-options) or `GOSS_LOGLEVEL`. As a library
it is your handler's, and goss neither reads nor interprets a level name.

## Record schema

Every message is a constant and every attribute key is `snake_case`, so records
can be matched and indexed on rather than parsed.

| Level | Message | Attributes |
|-------|---------|------------|
| `INFO` | `server listening` | `listen_addr` |
| `WARN` | `duplicate resource overwritten` | `resource_type`, `resource_id` |
| `WARN` | `empty configuration not written` | `path` |
| `DEBUG` | `command output` | `resource_id`, `command`, `stream` (`stdout` or `stderr`), `output` |
| `DEBUG` | `validation summary` (`json`) | `status` (`ok` or `fail`), `results_json` |
| `DEBUG` | `validation summary` (`rspecish`) | `status`, `total`, `failed`, `skipped`, `duration_seconds` |
| `DEBUG` | `request complete` | `client_addr`, `http_status`, and `response_body` only when the status is not 200 |
| `DEBUG` | `using configured output format` | `error` |
| `TRACE` | `request received` | `client_addr` |
| `TRACE` | `returning cached result` | `cache_key` |
| `TRACE` | `running validation for stale cache` | `cache_key` |
| `TRACE` | `validation result` | `outcome` (`success`, `fail` or `skip`), `resource_type`, `resource_id`, `property`, `expected`, `actual`, `duration_seconds`, and `result` under the `json` outputer |

`command output` emits one record per line, including a final empty one when the
output ends in a newline. It carries both identifiers because they differ
whenever a resource sets an explicit `exec`: `resource_id` is what the gossfile
called the resource, `command` is what ran. `expected` and `actual` are logged as native values,
so a handler decides how to render them.

One record does not follow this schema: errors reported by `http.Server` while
serving are forwarded at `ERROR` with whatever message the standard library
produced. goss does not parse that text into attributes.

## Embedding

Pass a logger to `util.NewConfig` and goss uses it everywhere, including the
subject commands it runs and the outputer that writes results:

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

c, err := util.NewConfig(
	util.WithSpecFile("goss.yaml"),
	util.WithLogger(logger),
)
if err != nil {
	return err
}

code, err := goss.Validate(c)
```

`util.Config` is also usable as a composite literal, in which case set `Logger`
directly. Either way, a nil logger means goss emits nothing: there is
deliberately no fallback to `slog.Default()`, because a library has no business
choosing where your records go. That silence is a guarantee, not an accident, so
goss also leaves `slog.SetDefault` and the standard `log` package alone.

The lower-level entry points take a logger too. `system.New` accepts
`system.WithLogger(logger)`, and `outputs.OutputConfig` has a `Logger` field.
`Validate`, `Serve` and the rest wire these up from the config for you.

## Rendering TRACE on your own handler

`util.ReplaceTraceLevel` has the signature of
`slog.HandlerOptions.ReplaceAttr`, so install it to make goss's trace records
render as the CLI renders them:

```go
slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level:       util.LevelTrace,
	ReplaceAttr: util.ReplaceTraceLevel,
}))
```

Note the `Level`, which is what a handler needs in order to be asked about a
record below `DEBUG` in the first place.

It only rewrites the built-in level attribute, and only when that attribute
holds `util.LevelTrace`, so it composes with a hook of your own:

```go
replace := func(groups []string, a slog.Attr) slog.Attr {
	return util.ReplaceTraceLevel(groups, scrub(groups, a))
}
```

If you emit your own trace records, `util.Trace` and `util.TraceContext` log at
`util.LevelTrace` and report the calling line as their source. Both accept a nil
logger.

## Sensitive fields

Four payloads that used to be interpolated into message text are now named
attributes: `output`, `actual`, `expected` and `response_body`. No record
changed level in that move, and all four remain at `DEBUG` or below. Still, a
named field in an indexed store is queryable and inherits that index's
retention, which is a wider exposure than the same bytes inside a message
string.

What can end up in a record:

- `output` is a subject command's stdout and stderr, verbatim.
- `expected` and `actual` are matcher values, which include fragments of
  vars-substituted gossfile content. The rendered gossfile itself is only ever
  written to stdout by `goss render`, never to a log record.
- `response_body` is a failing `goss serve` response, which contains the
  validation results.
- `resource_id`, `path` and parser diagnostics carry identifiers from your
  gossfile. `duplicate resource overwritten` is at `WARN`, so its identifier is
  visible at levels where nothing else here is.

None of this is redacted, and no attribute has a size cap. Filter what you do
not want to store in your own handler, either with `ReplaceAttr` or by keeping
goss at `INFO` or above. Composing a scrubber with `ReplaceTraceLevel` is shown
above.

## Migrating from earlier versions

Records are `slog` output now, so the `[LEVEL] ` prefix and the
`2006-01-02T15:04:05Z message` shape are gone, along with the prefix-matching
filter that implemented level selection. Timestamps from a CLI-built handler are
UTC.

For embedders:

- `util.Config.LogLevel` no longer exists. Set the level on the handler you pass
  to `util.WithLogger` instead. The field was removed rather than deprecated so
  that code setting it fails to compile, instead of building cleanly and
  logging nothing.
- Nothing reaches the standard `log` package any more, so a library caller that
  was capturing goss output through `log.SetOutput` sees nothing until it passes
  a logger.
- Calling the exported `GossConfig.Merge` directly no longer warns about
  duplicate resources. It has no config and therefore no logger; the operations
  that take a `*util.Config` still warn.
- `system.New` takes variadic options. Existing calls compile unchanged.
- The two deprecation notices for `--retry-timeout` and `--sleep` are printed
  straight to stderr, as before, and are not routed through a logger.

For CLI users:

- `--loglevel` now works under `render` and `autoadd`, where it was documented
  but had no effect.
- The `serve` request record no longer looks like `addr: status 200`. It is
  `request complete` with `client_addr` and `http_status` attributes, and the
  response body appears as `response_body` on failure only, without the old
  `" - "` prefix.
- `json` and `rspecish` summaries are `validation summary` records with
  attributes. The information they carried is preserved: the JSON document as
  `results_json`, rspecish's counts and timing as numbers.
- Trace records for validation results are `validation result` with named
  resource, outcome, value and duration attributes.
- Errors printed as goss exits still use the standard logger, so their timestamps
  are in local time and carry no level.
