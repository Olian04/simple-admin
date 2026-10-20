package ui

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Olian04/webui/pkg/webui"

	"github.com/Olian04/simple-admin/internal/backend"
	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// RecordArgs is the argument of every page about one record.
type RecordArgs struct {
	ID string `webui:"id"`
}

var segmentRE = regexp.MustCompile(`^[A-Za-z0-9._~-]+$`)

// pages declares every page of the app from the registry.
func (b *builder) pages() webui.Pages {
	var pages webui.Pages
	for _, be := range b.reg.Backends() {
		for _, rt := range b.reg.ResourcesFor(be.ID) {
			pages = append(pages, b.resourcePages(rt)...)
		}
	}
	return append(pages, b.credentialPages()...)
}

type paths struct{ list, detail, edit, new string }

func pathsFor(d resource.ResourceDescriptor) paths {
	base := "/" + d.Backend.ID + "/" + d.Name
	return paths{list: base, detail: base + "/{id}", edit: base + "/{id}/edit", new: base + "/new"}
}

// ordered lists a resource's fields for reading: the id first, then the
// writable fields in schema order, then the other read-only ones (timestamps).
// A discovered schema has no meaningful order of its own.
func ordered(d resource.ResourceDescriptor) []resource.Field {
	var id, writable, readOnly []resource.Field
	for _, f := range d.Schema.Fields {
		switch {
		case f.Name == d.IDField:
			id = append(id, f)
		case f.ReadOnly:
			readOnly = append(readOnly, f)
		default:
			writable = append(writable, f)
		}
	}
	return append(append(id, writable...), readOnly...)
}

// uniqueLabels gives each field a label no other field of the resource has.
func uniqueLabels(fields []resource.Field) map[string]string {
	out := make(map[string]string, len(fields))
	taken := map[string]bool{}
	for _, f := range fields {
		l := f.Label
		if l == "" || taken[strings.ToLower(l)] {
			l = f.Name
		}
		for taken[strings.ToLower(l)] {
			l += "_"
		}
		taken[strings.ToLower(l)] = true
		out[f.Name] = l
	}
	return out
}

func (b *builder) resourcePages(rt *resource.Runtime) webui.Pages {
	d := rt.Descriptor
	caps := d.Capabilities
	if !segmentRE.MatchString(d.Backend.ID) || !segmentRE.MatchString(d.Name) {
		b.log.Warn("resource skipped: id and name must be URL-safe", "backend", d.Backend.ID, "resource", d.Name)
		return nil
	}
	if !caps.List {
		return nil
	}
	p := pathsFor(d)
	labels := uniqueLabels(d.Schema.Fields)

	listID := webui.PageID[webui.NoArgs](p.list)
	detailID := webui.PageID[RecordArgs](p.detail)
	editID := webui.PageID[RecordArgs](p.edit)
	newID := webui.PageID[webui.NoArgs](p.new)

	var pages webui.Pages

	// List.
	var listBody webui.PageBody = b.table(rt, labels, detailID, caps.Read)
	if caps.Create {
		listBody = webui.Stack{
			webui.Toolbar[webui.NoArgs]{Left: []webui.ToolbarItem[webui.NoArgs]{
				webui.Link[webui.NoArgs, webui.NoArgs]{
					Label: "New " + singular(d.Label), Role: webui.RolePrimary, Page: newID,
					Args: func(context.Context, webui.NoArgs) webui.NoArgs { return webui.NoArgs{} },
				},
			}},
			listBody,
		}
	}
	pages = append(pages, webui.Page[webui.NoArgs]{
		Path: listID,
		Nav:  webui.Nav{Label: d.Label, Icon: "table-list", Section: d.Backend.Label},
		Body: listBody,
	})

	// Detail.
	if caps.Read {
		var items []webui.ToolbarItem[RecordArgs]
		same := func(_ context.Context, a RecordArgs) RecordArgs { return a }
		if caps.Update {
			items = append(items, webui.Link[RecordArgs, RecordArgs]{Label: "Edit", Role: webui.RolePrimary, Page: editID, Args: same})
		}
		if caps.Delete {
			items = append(items, webui.Action[RecordArgs]{
				Label:   "Delete",
				Role:    webui.RoleDestructive,
				Confirm: "Delete this " + strings.ToLower(singular(d.Label)) + "? This cannot be undone.",
				Run: func(ctx context.Context, a RecordArgs) (webui.Outcome, error) {
					err := rt.Service.Delete(b.scope(ctx, d.Backend.ID), a.ID)
					if err != nil {
						return b.failure(d, err, nil)
					}
					return webui.Success(singular(d.Label) + " deleted").Then(webui.Open(ctx, listID, webui.NoArgs{})), nil
				},
			})
		}
		var detailBody webui.PageBody = webui.Form[rec]{
			Title:  singular(d.Label),
			Load:   b.loader(rt),
			Fields: b.formFields(rt, labels, false, false),
		}
		if len(items) > 0 {
			detailBody = webui.Stack{webui.Toolbar[RecordArgs]{Left: items}, detailBody}
		}
		pages = append(pages, webui.Page[RecordArgs]{Path: detailID, Body: detailBody})
	}

	// Create.
	if caps.Create {
		fields, byName := b.editFields(rt, labels)
		pages = append(pages, webui.Page[webui.NoArgs]{
			Path: newID,
			Body: webui.Form[rec]{
				Title:  "New " + strings.ToLower(singular(d.Label)),
				Load:   func(context.Context) (rec, error) { return rec{}, nil },
				Fields: fields,
				Submit: webui.Action[rec]{
					Label: "Create",
					Run: func(ctx context.Context, r rec) (webui.Outcome, error) {
						sctx := b.scope(ctx, d.Backend.ID)
						created, err := rt.Service.Create(sctx, body(d, r))
						if err != nil {
							return b.failure(d, err, byName)
						}
						msg := singular(d.Label) + " created"
						if id := str(created[d.IDField]); id != "" && caps.Read {
							return webui.Success(msg).Then(webui.Open(ctx, detailID, RecordArgs{ID: id})), nil
						}
						return webui.Success(msg).Then(webui.Open(ctx, listID, webui.NoArgs{})), nil
					},
				},
			},
		})
	}

	// Edit.
	if caps.Update && caps.Read {
		fields, byName := b.editFields(rt, labels)
		pages = append(pages, webui.Page[RecordArgs]{
			Path: editID,
			Body: webui.Form[rec]{
				Title:  "Edit " + strings.ToLower(singular(d.Label)),
				Load:   b.loader(rt),
				Fields: fields,
				Submit: webui.Action[rec]{
					Label: "Save",
					Run: func(ctx context.Context, r rec) (webui.Outcome, error) {
						id := webui.ArgsOf[RecordArgs](ctx).ID
						sctx := b.scope(ctx, d.Backend.ID)
						_, err := rt.Service.Update(sctx, id, body(d, r))
						if err != nil {
							return b.failure(d, err, byName)
						}
						return webui.Success(singular(d.Label) + " updated").Then(webui.Open(ctx, detailID, RecordArgs{ID: id})), nil
					},
				},
			},
		})
	}

	return pages
}

// singular makes a heading of a resource label: "Posts" → "Post".
func singular(label string) string {
	if strings.HasSuffix(label, "ies") && len(label) > 3 {
		return label[:len(label)-3] + "y"
	}
	if strings.HasSuffix(label, "s") && !strings.HasSuffix(label, "ss") && len(label) > 1 {
		return label[:len(label)-1]
	}
	return label
}

// loader loads the record the page's ID names.
func (b *builder) loader(rt *resource.Runtime) func(context.Context) (rec, error) {
	d := rt.Descriptor
	return func(ctx context.Context) (rec, error) {
		id := webui.ArgsOf[RecordArgs](ctx).ID
		r, err := rt.Service.Get(b.scope(ctx, d.Backend.ID), id)
		if err != nil {
			return nil, b.readErr(ctx, d, err)
		}
		if r == nil {
			r = rec{}
		}
		return r, nil
	}
}

// table lists the resource's scalar fields as a page of the upstream list.
func (b *builder) table(rt *resource.Runtime, labels map[string]string, detail webui.PageID[RecordArgs], canOpen bool) webui.Table[rec] {
	d := rt.Descriptor
	var cols []webui.Accessor[rec]
	byLabel := map[string]string{}
	for _, f := range ordered(d) {
		if f.Kind == resource.KindArray || f.Kind == resource.KindObject {
			continue
		}
		byLabel[labels[f.Name]] = f.Name
		cols = append(cols, b.accessor(rt, f, labels[f.Name], accOpts{column: true, first: len(cols) == 0}))
	}
	t := webui.Table[rec]{
		Title:    d.Label,
		PageSize: b.pageSize,
		Columns:  cols,
		Load: func(ctx context.Context, q webui.Query) (webui.Window[rec], error) {
			per := q.Limit
			if per < 1 {
				per = b.pageSize
			}
			res, err := rt.Service.List(b.scope(ctx, d.Backend.ID), resource.ListParams{
				Page: q.Offset/per + 1, PerPage: per, Sort: byLabel[q.Sort], Desc: q.Desc,
			})
			if err != nil {
				return webui.Window[rec]{}, b.readErr(ctx, d, err)
			}
			return webui.Window[rec]{Items: res.Items, Total: res.TotalItems}, nil
		},
	}
	if canOpen {
		t.RowClick = webui.Link[rec, RecordArgs]{
			Page: detail,
			Args: func(_ context.Context, r rec) RecordArgs { return RecordArgs{ID: str(r[d.IDField])} },
		}
	}
	return t
}

// formFields lists every field of the resource, read-only unless editable.
func (b *builder) formFields(rt *resource.Runtime, labels map[string]string, editable, skipReadOnly bool) []webui.Accessor[rec] {
	var out []webui.Accessor[rec]
	for _, f := range ordered(rt.Descriptor) {
		if skipReadOnly && f.ReadOnly {
			continue
		}
		out = append(out, b.accessor(rt, f, labels[f.Name], accOpts{editable: editable}))
	}
	return out
}

// editFields are the inputs of the create and edit forms, with the accessor of
// each field by name so a validation error can point at it.
func (b *builder) editFields(rt *resource.Runtime, labels map[string]string) ([]webui.Accessor[rec], map[string]webui.Accessor[rec]) {
	var out []webui.Accessor[rec]
	byName := map[string]webui.Accessor[rec]{}
	for _, f := range rt.Descriptor.Schema.Fields {
		if f.ReadOnly {
			continue
		}
		acc := b.accessor(rt, f, labels[f.Name], accOpts{editable: true})
		byName[f.Name] = acc
		out = append(out, acc)
	}
	return out, byName
}

// body is the request body of a write: the writable fields of the record.
func body(d resource.ResourceDescriptor, r rec) resource.Record {
	out := resource.Record{}
	for _, f := range d.Schema.Fields {
		if f.ReadOnly {
			continue
		}
		if v, ok := r[f.Name]; ok && v != nil {
			out[f.Name] = v
		}
	}
	return out
}

// failure turns a write error into what the visitor should see.
func (b *builder) failure(d resource.ResourceDescriptor, err error, fields map[string]webui.Accessor[rec]) (webui.Outcome, error) {
	var verr *resource.ValidationError
	if errors.As(err, &verr) {
		var rejects []webui.FieldError[rec]
		var all []string
		general := false
		for name, msg := range verr.Fields {
			all = append(all, name+": "+msg)
			if acc, ok := fields[name]; ok {
				rejects = append(rejects, webui.Field[rec](acc, msg))
			} else {
				general = true
			}
		}
		if general || len(rejects) == 0 {
			return webui.Failure(strings.Join(all, "; ")), nil
		}
		return webui.Reject(rejects...), nil
	}
	var up *backend.Error
	switch {
	case errors.Is(err, backend.ErrUnauthorized):
		return webui.Failure(d.Backend.Label + " refused the credential. Set a valid one on the Credentials page."), nil
	case errors.Is(err, backend.ErrNotFound):
		return webui.Failure("Not found: it may have been deleted."), nil
	case errors.As(err, &up):
		return webui.Failure(fmt.Sprintf("%s answered HTTP %d: %s", d.Backend.Label, up.Status, truncate(up.Body, 200))), nil
	default:
		return webui.Outcome{}, err
	}
}
