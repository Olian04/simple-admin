# webui gaps found while building simple-admin

Checked against [webui](https://github.com/Olian04/webui) **v1.0.0-19**. Each is a use case for an issue at <https://github.com/Olian04/webui/issues>; "Workaround" is what `internal/ui` does today.

Closed by webui while this was written, and now used here: `Select.LoadOptions` and `NoSort`/`NoFilter` (v1.0.0-15), `Toolbar` (v1.0.0-16), `String.Multiline`, `String.Secret`, `ValidatePattern`, `Refusal`, recompiling into the same metrics registry, `OptionalInt`/`OptionalFloat`/`OptionalSlider` and `Bool` (v1.0.0-17/18), `Action.Confirm`, `Email`/`Website`/`ColorPicker`, `DateOnly` and `Help` (v1.0.0-19).

| # | Gap | Why simple-admin needs it | Workaround today | Impact |
|---|---|---|---|---|
| 1 | **No code / JSON / markdown editor as a form field.** `Editor` is its own leaf with its own submit; `Multiline` is a plain text area. | Syntax-highlighted editing of `object`/`array` and code fields. | `Multiline` text; JSON is parsed on submit and checked by the JSON Schema. | Medium |
| 2 | **No free-form list / tags input** (`Select Multi` needs a known option set). | `array` of strings. | Comma separated text. | Medium |
| 3 | **No typeahead / search for a `Select`.** `LoadOptions` runs on every request that draws, checks, filters or exports the select, and is not given a query. | A relation to a resource with thousands of records. | First 200 records, cached for 10 s; a record outside them is refused when the form is saved unchanged. | Medium |
| 4 | **A `Load` `Refusal` cannot set the page status.** Only a `Guard` can answer 404/401; a panel's `Load` says it in the panel and the page stays 200. | A missing record as a real 404 without a second upstream read in a `Guard`. | The panel says "There is no such post." with HTTP 200. | Low |
| 5 | **A cell cannot link to another page** (`Link` is only a row click; `URL` takes a string). | A relation column that opens the related record. | Shown as a badge with the record's label; the row click opens the row itself. | Low |
| 6 | **A crumb comes from the path, not from a page without arguments.** `Page.Crumb` renames a page's own crumb, but there is no page at `/ex`, so the backend id shows as "Ex". | Backend label in the trail. | Backend label is the sidebar section. | Low |
| 7 | **No tel input without the country-code splitting of `Phone`; an optional `ColorPicker` cannot be empty.** | `format: color` on optional fields. | An optional color stays plain text. | Low |
| 8 | **No file upload.** | `file` / `image` fields (never implemented in the old frontend either). | Not supported. | Low |
