package ui

import (
	"context"

	"github.com/Olian04/webui/pkg/webui"
)

// credentialsID is where visitors give the keys of the backends that want one.
const credentialsID webui.PageID[webui.NoArgs] = "/credentials"

// credModel is what the credentials form edits: one input per backend that
// takes a key, and a read-only line saying whether one is set.
type credModel struct {
	set    map[string]bool
	values map[string]string
}

// credentialPages is the page where a visitor gives the key for each backend
// that wants one. It exists only when some backend does.
func (b *builder) credentialPages() webui.Pages {
	type be struct {
		id, label string
	}
	var needs []be
	seen := map[string]bool{}
	for _, backend := range b.reg.Backends() {
		if seen[backend.ID] {
			continue
		}
		for _, rt := range b.reg.ResourcesFor(backend.ID) {
			if rt.Descriptor.AuthScheme.Type == "apikey" {
				seen[backend.ID] = true
				needs = append(needs, be{backend.ID, backend.Label})
				break
			}
		}
	}
	if len(needs) == 0 {
		return nil
	}

	var fields []webui.Accessor[credModel]
	for _, n := range needs {
		id := n.id
		fields = append(fields, webui.String[credModel]{
			Label:  n.label + " API key",
			Secret: true,
			// A Secret is never read back: the page only learns whether there is one.
			Load: func(m credModel) string {
				if m.set[id] {
					return "set"
				}
				return ""
			},
			Placeholder: "API key",
			Store:       func(m *credModel, v string) { m.values[id] = v },
		})
	}

	form := webui.Form[credModel]{
		Title: "Credentials",
		Desc:  "Keys are kept in memory for this browser session only, and sent to the backend with each request. Leave a field empty to keep the current key.",
		Load: func(ctx context.Context) (credModel, error) {
			m := credModel{set: map[string]bool{}, values: map[string]string{}}
			for _, n := range needs {
				m.set[n.id] = b.sessions.Credential(ctx, n.id) != ""
			}
			return m, nil
		},
		Fields: fields,
		Submit: webui.Action[credModel]{
			Label: "Save",
			Run: func(ctx context.Context, m credModel) (webui.Outcome, error) {
				for id, v := range m.values {
					if v != "" {
						b.sessions.SetCredential(ctx, id, v)
					}
				}
				return webui.Success("Credentials saved").Stay(), nil
			},
		},
	}
	toolbar := webui.Toolbar[webui.NoArgs]{Left: []webui.ToolbarItem[webui.NoArgs]{
		webui.Action[webui.NoArgs]{
			Label: "Clear credentials", Role: webui.RoleDestructive,
			Run: func(ctx context.Context, _ webui.NoArgs) (webui.Outcome, error) {
				b.sessions.ClearCredentials(ctx)
				return webui.Success("Credentials cleared"), nil
			},
		},
	}}
	return webui.Pages{webui.Page[webui.NoArgs]{
		Path: credentialsID,
		Nav:  webui.Nav{Label: "Credentials", Icon: "key", Section: "Settings"},
		Body: webui.Stack{toolbar, form},
	}}
}
