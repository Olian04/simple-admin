# Step p2-12 — High-value widgets (color/json/tags/markdown/code)

Status: `done`

## Sub-tasks

- [x] Templ widgets: `widgetColor` (native), `widgetJSON` (textarea + Format), `widgetTags` (chips + hidden), `widgetMarkdown` (EasyMDE host), `widgetCode` (Monaco host); map them in `builtinWidgets` (removed from stubs)
- [x] Islands (no external dep): `json.js` (format/validate), `tags.js` (chip input ↔ CSV hidden)
- [x] Islands (CDN): `markdown.js` (EasyMDE from `data-easymde-cdn`), `code.js` (Monaco from `data-monaco-cdn` + worker shim); load all in layout
- [x] `app.css`: `.sa-code-editor` height, `.sa-color`
- [x] CSP: `securityHeaders(next, ui)` allowlists configured CDN origins (script/style/img/font/connect) + `style 'unsafe-inline'` + `worker-src blob: data:` (Monaco); `originList` helper
- [x] Fixed revive `min` shadow in poller.go

## Verification

- generate/build/vet/golangci(0)/tests pass
- Smoke: code field renders `data-sa-code` + `sa-code-editor`; `/ui` CSP header includes jsdelivr + worker-src
- NOTE: actual Monaco/EasyMDE editor mount requires a browser (JS) — not covered by HTTP smoke; manual check pending

## Commit

- (next)
