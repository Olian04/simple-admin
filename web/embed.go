// Package web embeds the admin UI's static assets (vendored Bulma + HTMX,
// handwritten app.css, and JS islands). Embedding lives here because go:embed
// cannot reference parent directories from internal/.
package web

import "embed"

//go:embed static
var Static embed.FS
