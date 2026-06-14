package views

import (
	"strconv"
	"strings"

	"github.com/a-h/templ"
)

// uiPath joins non-empty segments under /ui (e.g. uiPath("ex","posts","42") ->
// "/ui/ex/posts/42"; trailing "" segments are dropped).
func uiPath(parts ...string) string {
	segs := make([]string, 0, len(parts)+1)
	segs = append(segs, "ui")
	for _, p := range parts {
		if p != "" {
			segs = append(segs, p)
		}
	}
	return "/" + strings.Join(segs, "/")
}

func uiURL(parts ...string) templ.SafeURL {
	return templ.SafeURL(uiPath(parts...))
}

func colspan(n int) string { return strconv.Itoa(n) }

// pageLabel summarizes the current page position.
func pageLabel(page, totalPages, totalItems int) string {
	return "Page " + strconv.Itoa(page) + " / " + strconv.Itoa(totalPages) +
		" · " + strconv.Itoa(totalItems) + " items"
}

// schemeLabel describes a backend's auth scheme for the credentials page.
func schemeLabel(b CredBackend) string {
	switch b.AuthType {
	case "apikey":
		return "API key (" + b.In + ": " + b.Name + ")"
	case "bearer":
		return "Bearer token (not yet supported)"
	default:
		return b.AuthType
	}
}
