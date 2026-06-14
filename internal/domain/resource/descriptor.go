// Package resource holds the canonical, source-agnostic admin domain model.
//
// It imports only the standard library and its own ports. OpenAPI, HTTP, and
// config types never leak in here — sources (manual config, OpenAPI discovery)
// translate their inputs into the types defined in this package.
package resource

// FieldKind classifies how a field renders and validates.
type FieldKind string

const (
	KindString  FieldKind = "string"
	KindNumber  FieldKind = "number"
	KindInteger FieldKind = "integer"
	KindBoolean FieldKind = "boolean"
	KindEnum    FieldKind = "enum"
	KindArray   FieldKind = "array"
	KindObject  FieldKind = "object"
)

// Field is one property of a resource schema, kept in declared order.
type Field struct {
	Name      string
	Label     string
	Kind      FieldKind
	Format    string // email|uri|date|date-time|password|color|...
	Required  bool
	ReadOnly  bool
	Enum      []string
	Min       *float64
	Max       *float64
	MaxLength *int
	Pattern   string
	Items     *Field         // element schema for arrays
	Custom    *FieldTemplate // optional widget override (config-supplied)
}

// FieldTemplate is a resolved widget binding for a field (built-in id or a
// user-supplied template file) plus its configured args.
type FieldTemplate struct {
	TemplateID   string
	TemplateFile string
	Args         map[string]TemplateArg
}

// TemplateArg is a static value or a JSONPath selector resolved against the
// record at render time.
type TemplateArg struct {
	Static   any
	JSONPath string
}

// Schema is the resource's JSON Schema: raw bytes for validation plus the
// ordered fields derived from it for UI rendering.
type Schema struct {
	Raw    []byte
	Fields []Field
}

// Capabilities reflect which CRUD verbs the backend implements for a resource.
type Capabilities struct {
	List   bool
	Read   bool
	Create bool
	Update bool
	Delete bool
	// UpdateVerb is the HTTP method used for updates: "PUT" or "PATCH".
	UpdateVerb string
}

// AuthScheme is auth metadata for upstream calls. It never carries a secret —
// credentials are supplied per-request (browser → passthrough).
type AuthScheme struct {
	Type string // none|apikey|bearer|basic|oauth
	In   string // header|query
	Name string
}

// CRUDPaths are the upstream paths, resolved against the backend base URL.
type CRUDPaths struct {
	Collection string
	Item       string // contains the {id} placeholder
}

// Backend identifies an upstream resource owner.
type Backend struct {
	ID      string
	Label   string
	BaseURL string
}

// ResourceDescriptor is the canonical model the UI and transport speak.
type ResourceDescriptor struct {
	Backend      Backend
	Name         string
	Label        string
	Schema       Schema
	Capabilities Capabilities
	AuthScheme   AuthScheme
	IDField      string
	CRUDPaths    CRUDPaths
}

// Key is the registry key for a descriptor: "<backendID>/<name>".
func (d ResourceDescriptor) Key() string { return d.Backend.ID + "/" + d.Name }
