package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"

	"github.com/Olian04/simple-admin/internal/backend"
	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/views"
	"github.com/Olian04/simple-admin/internal/views/form"
)

// NewForm renders an empty create form.
func (h *UIHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	rt, ok := h.lookup(w, r)
	if !ok {
		return
	}
	if !rt.Descriptor.Capabilities.Create {
		h.renderStatus(w, r, http.StatusNotFound, rt.Descriptor, "Not allowed", "Create is not supported for this resource")
		return
	}
	vm := h.buildFormView(rt.Descriptor, "create", "", nil, nil)
	data := h.pageData(rt.Descriptor.Backend.ID, rt.Descriptor.Name, form.Form(vm))
	h.render(w, r, views.Page(data))
}

// EditForm renders an edit form seeded with the record.
func (h *UIHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	rt, ok := h.lookup(w, r)
	if !ok {
		return
	}
	if !rt.Descriptor.Capabilities.Update {
		h.renderStatus(w, r, http.StatusNotFound, rt.Descriptor, "Not allowed", "Update is not supported for this resource")
		return
	}
	id := r.PathValue("id")
	rec, err := rt.Service.Get(r.Context(), id)
	if errors.Is(err, backend.ErrNotFound) {
		h.renderStatus(w, r, http.StatusNotFound, rt.Descriptor, "Not found", "No "+rt.Descriptor.Label+" with id "+id)
		return
	}
	if err != nil {
		h.renderError(w, r, rt.Descriptor, "Failed to load record", err)
		return
	}
	vm := h.buildFormView(rt.Descriptor, "edit", id, rec, nil)
	data := h.pageData(rt.Descriptor.Backend.ID, rt.Descriptor.Name, form.Form(vm))
	h.render(w, r, views.Page(data))
}

// Create validates and creates a record.
func (h *UIHandler) Create(w http.ResponseWriter, r *http.Request) {
	rt, ok := h.lookup(w, r)
	if !ok {
		return
	}
	rec, err := parseForm(r, rt.Descriptor)
	if err != nil {
		h.formError(w, r, rt.Descriptor, "create", "", rec, err)
		return
	}
	created, err := rt.Service.Create(r.Context(), rec)
	if err != nil {
		h.handleWriteError(w, r, rt.Descriptor, "create", "", rec, err)
		return
	}
	h.redirectToRecord(w, rt.Descriptor, created)
}

// Update validates and updates a record.
func (h *UIHandler) Update(w http.ResponseWriter, r *http.Request) {
	rt, ok := h.lookup(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	rec, err := parseForm(r, rt.Descriptor)
	if err != nil {
		h.formError(w, r, rt.Descriptor, "edit", id, rec, err)
		return
	}
	if _, err := rt.Service.Update(r.Context(), id, rec); err != nil {
		h.handleWriteError(w, r, rt.Descriptor, "edit", id, rec, err)
		return
	}
	w.Header().Set("HX-Redirect", "/ui/"+rt.Descriptor.Backend.ID+"/"+rt.Descriptor.Name+"/"+id)
	w.WriteHeader(http.StatusOK)
}

// handleWriteError renders validation errors (422) or an upstream error (502)
// back into the form for an HTMX swap.
func (h *UIHandler) handleWriteError(w http.ResponseWriter, r *http.Request, d resource.ResourceDescriptor, mode, id string, rec resource.Record, err error) {
	var verr *resource.ValidationError
	if errors.As(err, &verr) {
		vm := h.buildFormView(d, mode, id, rec, verr.Fields)
		w.WriteHeader(http.StatusUnprocessableEntity)
		h.render(w, r, form.FormBody(vm))
		return
	}
	h.formError(w, r, d, mode, id, rec, err)
}

func (h *UIHandler) formError(w http.ResponseWriter, r *http.Request, d resource.ResourceDescriptor, mode, id string, rec resource.Record, err error) {
	vm := h.buildFormView(d, mode, id, rec, nil)
	vm.FormError = err.Error()
	w.WriteHeader(http.StatusBadGateway)
	h.render(w, r, form.FormBody(vm))
}

func (h *UIHandler) redirectToRecord(w http.ResponseWriter, d resource.ResourceDescriptor, rec resource.Record) {
	target := "/ui/" + d.Backend.ID + "/" + d.Name
	if id := views.DisplayValue(rec[d.IDField]); id != "" {
		target += "/" + id
	}
	w.Header().Set("HX-Redirect", target)
	w.WriteHeader(http.StatusOK)
}

// parseForm reads form values into a Record, coercing types per field kind.
// Read-only fields are skipped; empty optional scalars are omitted.
func parseForm(r *http.Request, d resource.ResourceDescriptor) (resource.Record, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("parse form: %w", err)
	}
	rec := resource.Record{}
	for _, f := range d.Schema.Fields {
		if f.ReadOnly {
			continue
		}
		switch f.Kind {
		case resource.KindBoolean:
			rec[f.Name] = r.PostForm.Get(f.Name) == "true"
		case resource.KindNumber, resource.KindInteger:
			v := strings.TrimSpace(r.PostForm.Get(f.Name))
			if v == "" {
				continue
			}
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("%s must be a number", f.Label)
			}
			rec[f.Name] = n
		case resource.KindArray, resource.KindObject:
			v := strings.TrimSpace(r.PostForm.Get(f.Name))
			if v == "" {
				continue
			}
			var parsed any
			if json.Unmarshal([]byte(v), &parsed) == nil {
				rec[f.Name] = parsed
			} else {
				rec[f.Name] = v
			}
		default:
			v := r.PostForm.Get(f.Name)
			if v == "" {
				continue
			}
			rec[f.Name] = v
		}
	}
	return rec, nil
}

func (h *UIHandler) buildFormView(d resource.ResourceDescriptor, mode, id string, data resource.Record, errs map[string]string) form.FormView {
	known := h.resourceNames(d.Backend.ID)
	ff := form.FormFields(d.Schema.Fields, data, errs)
	fields := make([]form.FormFieldView, 0, len(ff))
	for _, f := range ff {
		fv := form.FormFieldView{
			Name:      f.Name,
			Label:     f.Label,
			Kind:      string(f.Kind),
			InputType: form.InputType(f.Field),
			Required:  f.Required,
			Error:     f.Error,
			Enum:      f.Enum,
		}
		if f.Kind == resource.KindBoolean {
			fv.Checked = isTruthy(f.Value)
		} else {
			fv.Value = views.ValueString(f.Value)
		}
		if c, ok := h.renderWidget(d, f, data, known); ok {
			fv.Widget = c
		}
		fields = append(fields, fv)
	}

	action := "/ui/" + d.Backend.ID + "/" + d.Name
	cancel := action
	usePut := false
	if mode == "edit" {
		action += "/" + id
		cancel = action
		usePut = true
	}
	return form.FormView{
		Backend:   d.Backend,
		Name:      d.Name,
		Label:     d.Label,
		Mode:      mode,
		ID:        id,
		Action:    action,
		UsePut:    usePut,
		CancelURL: cancel,
		Fields:    fields,
	}
}

// DeleteConfirm renders the delete-confirmation modal (loaded into #sa-modal).
func (h *UIHandler) DeleteConfirm(w http.ResponseWriter, r *http.Request) {
	rt, ok := h.lookup(w, r)
	if !ok {
		return
	}
	if !rt.Descriptor.Capabilities.Delete {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	d := rt.Descriptor
	h.render(w, r, views.DeleteConfirm(views.DeleteView{
		Backend: d.Backend, Name: d.Name, Label: d.Label, ID: r.PathValue("id"),
	}))
}

// Delete removes a record; on success it redirects (HX-Redirect) to the list.
func (h *UIHandler) Delete(w http.ResponseWriter, r *http.Request) {
	rt, ok := h.lookup(w, r)
	if !ok {
		return
	}
	d := rt.Descriptor
	if !d.Capabilities.Delete {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	id := r.PathValue("id")
	if err := rt.Service.Delete(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		h.render(w, r, views.DeleteConfirm(views.DeleteView{
			Backend: d.Backend, Name: d.Name, Label: d.Label, ID: id, Error: err.Error(),
		}))
		return
	}
	w.Header().Set("HX-Redirect", "/ui/"+d.Backend.ID+"/"+d.Name)
	w.WriteHeader(http.StatusOK)
}

func (h *UIHandler) resourceNames(backendID string) []string {
	rts := h.registry.ResourcesFor(backendID)
	names := make([]string, 0, len(rts))
	for _, rt := range rts {
		names = append(names, rt.Descriptor.Name)
	}
	return names
}

// renderWidget resolves a field's custom widget (explicit config or convention
// auto-select) and renders it to HTML. Returns ok=false to fall back to the
// built-in input.
func (h *UIHandler) renderWidget(d resource.ResourceDescriptor, f form.FormField, data resource.Record, known []string) (templ.Component, bool) {
	if h.widgets == nil {
		return nil, false
	}
	var widgetID, optionsURL string
	var args map[string]any

	if f.Custom != nil {
		if f.Custom.TemplateFile != "" {
			widgetID = f.Custom.TemplateFile
		} else {
			widgetID = f.Custom.TemplateID
		}
		args = form.ResolveArgs(f.Custom.Args, data)
	} else if id, target, ok := form.AutoSelectWidget(f.Field, known); ok {
		widgetID = id
		if id == "relation" && target != "" {
			optionsURL = "/ui/" + d.Backend.ID + "/" + target + "/options"
		}
	}

	if widgetID == "" || !h.widgets.Has(widgetID) {
		return nil, false
	}

	switch widgetID {
	case "relation":
		if optionsURL == "" {
			if res, _ := args["resource"].(string); res != "" {
				optionsURL = "/ui/" + d.Backend.ID + "/" + res + "/options"
			}
		}
	case "datetime":
		if args == nil {
			args = map[string]any{}
		}
		if _, ok := args["type"]; !ok {
			mode, _ := args["mode"].(string)
			if mode == "" {
				mode = modeFromFormat(f.Format)
			}
			args["type"] = form.DatetimeInputType(mode)
		}
	}

	return h.widgets.Component(widgetID, form.FieldContext{
		Name:       f.Name,
		Label:      f.Label,
		InputID:    "f_" + f.Name,
		Value:      views.ValueString(f.Value),
		Required:   f.Required,
		Args:       args,
		OptionsURL: optionsURL,
	})
}

func modeFromFormat(format string) string {
	switch format {
	case "date":
		return "date"
	case "time":
		return "time"
	default:
		return "datetime"
	}
}

func isTruthy(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true" || val == "✓"
	default:
		return false
	}
}
