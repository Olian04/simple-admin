package ui_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/observability/metrics"
	"github.com/Olian04/simple-admin/internal/ui"
)

func TestListShowsRowsNavAndNewLink(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	rec := get(e.h, "/ui/ex/posts")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	contains(t, rec,
		"Hello", "World", // rows
		">Authors<", // nav entry of another resource
		"New Post",  // toolbar link
		">Ada<",     // relation shown by label, not id
		"Yes",       // boolean
		`href="/ui/ex/posts/p1`,
	)
	if strings.Contains(rec.Body.String(), "filter") {
		t.Error("columns must not offer filters: the upstream contract has none")
	}
}

func TestListSortAndPageGoUpstream(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	rec := get(e.h, "/ui/ex/posts?posts.sort=title&posts.desc=true&posts.offset=25")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	p := e.posts.lastList
	if p.Sort != "title" || !p.Desc || p.Page != 2 || p.PerPage != 25 {
		t.Fatalf("upstream params wrong: %+v", p)
	}
}

func TestDetailShowsRecordWithEditAndDelete(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	rec := get(e.h, "/ui/ex/posts/p1")
	contains(t, rec, "Hello", "x, y", "/ui/ex/posts/p1/edit", "Delete this post? This cannot be undone.")
	// A missing record says so, and shows nothing of another record.
	missing := get(e.h, "/ui/ex/posts/nope")
	contains(t, missing, "There is no such post")
	if strings.Contains(missing.Body.String(), "Hello") {
		t.Error("unknown record shows another record")
	}
}

func TestCreateWritesOnlyTypedWritableFields(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	form := url.Values{
		"_leaf": {"p"}, "f0": {"New one"}, "f2": {"draft"}, "f3": {"2"},
		"f4": {"true"}, "f5": {"a2"}, "f6": {"go, web"}, "f7": {`{"k":1}`},
	}
	rec := post(e.h, "/ui/ex/posts/new", form)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d: %s", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); !strings.HasSuffix(loc, "/ui/ex/posts/n1") {
		t.Fatalf("should land on the new record, got %q", loc)
	}
	b := e.posts.lastBody
	if b["title"] != "New one" || b["status"] != "draft" || b["featured"] != true || b["author"] != "a2" {
		t.Fatalf("body wrong: %#v", b)
	}
	if tags, _ := b["tags"].([]any); len(tags) != 2 || tags[0] != "go" || tags[1] != "web" {
		t.Fatalf("tags wrong: %#v", b["tags"])
	}
	if meta, _ := b["meta"].(map[string]any); meta["k"] != 1.0 {
		t.Fatalf("meta wrong: %#v", b["meta"])
	}
	if _, has := b["body"]; has {
		t.Error("an empty optional string must be left out")
	}
	if _, has := b["id"]; has {
		t.Error("read-only fields must not be written")
	}
}

func TestCreateRejectsWhatTheSchemaRefusesBesideTheField(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	// meta must be an object; webui cannot know that, the JSON Schema does.
	rec := post(e.h, "/ui/ex/posts/new", url.Values{"_leaf": {"p"}, "f0": {"T"}, "f7": {"not json"}})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", rec.Code)
	}
	contains(t, rec, "not json") // what was typed is kept
	if len(e.posts.rows) != 2 {
		t.Error("nothing may be written when validation fails")
	}
}

func TestCreateRequiresRequiredFieldsInThePage(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	rec := post(e.h, "/ui/ex/posts/new", url.Values{"_leaf": {"p"}, "f0": {""}})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", rec.Code)
	}
}

func TestEditSendsOnlyWritableFieldsAndReturnsToTheRecord(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	rec := post(e.h, "/ui/ex/posts/p1/edit", url.Values{
		"_leaf": {"p"}, "f0": {"Renamed"}, "f2": {"archived"}, "f3": {"3"}, "f4": {"false"}, "f5": {"a1"}, "f6": {"x, y"}, "f7": {""},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d: %s", rec.Code, rec.Body.String())
	}
	b := e.posts.lastBody
	if b["title"] != "Renamed" || b["status"] != "archived" || b["featured"] != false {
		t.Fatalf("body wrong: %#v", b)
	}
	if _, has := b["id"]; has {
		t.Error("id must not be sent")
	}
	if e.posts.rows[0]["title"] != "Renamed" {
		t.Error("record not updated upstream")
	}
}

func TestDeleteAsksInTheButtonsDialogThenDeletes(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	contains(t, get(e.h, "/ui/ex/posts/p2"), "This cannot be undone.")
	if len(e.posts.rows) != 2 {
		t.Fatal("viewing the record must not delete")
	}
	rec := post(e.h, "/ui/ex/posts/p2", url.Values{"_leaf": {"p.0"}, "_act": {"1"}})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(e.posts.rows) != 1 || e.posts.rows[0]["id"] != "p1" {
		t.Fatalf("p2 not deleted: %v", e.posts.rows)
	}
}

func TestCapabilitiesDecideWhichPagesExist(t *testing.T) {
	e := newEnv(t, resourceNone())
	ro := runtime(t, e.posts, "posts", "Posts", postSchema, resourceNone(), resource.Capabilities{List: true, Read: true})
	e.reg.Replace([]*resource.Runtime{ro})
	e.build(t)
	rec := get(e.h, "/ui/ex/posts")
	if strings.Contains(rec.Body.String(), "New Post") {
		t.Error("no create capability: no New link")
	}
	for _, p := range []string{"/ui/ex/posts/p1/edit"} {
		if c := get(e.h, p).Code; c == http.StatusOK {
			t.Errorf("%s should not exist, got 200", p)
		}
	}
	// "new" is then just a record id, and there is no create form.
	if strings.Contains(get(e.h, "/ui/ex/posts/new").Body.String(), ">Create<") {
		t.Error("a resource that cannot create must not offer a create form")
	}
	if strings.Contains(get(e.h, "/ui/ex/posts/p1").Body.String(), "/edit") {
		t.Error("read-only detail must not link to edit")
	}
}

func TestCredentialsAreAskedForWhenTheBackendRefusesAndAreSentUpstream(t *testing.T) {
	e := newEnv(t, resource.AuthScheme{Type: "apikey", In: "header", Name: "X-API-Key"})
	e.posts.needCred = "secret"
	e.authors.needCred = "secret"
	e.build(t)

	first := get(e.h, "/ui/ex/posts")
	var session *http.Cookie
	for _, c := range first.Result().Cookies() {
		if c.Name == "sa_session" {
			session = c
		}
	}
	if session == nil || !session.HttpOnly {
		t.Fatalf("no HttpOnly session cookie: %v", first.Result().Cookies())
	}
	// The panel tells the visitor what to do, with a link to the page.
	contains(t, get(e.h, "/ui/ex/posts", session), "needs a credential", `href="/ui/credentials`)

	saved := post(e.h, "/ui/credentials", url.Values{"_leaf": {"p.1"}, "f0": {"secret"}}, session)
	if saved.Code != http.StatusSeeOther {
		t.Fatalf("saving credentials: want 303, got %d: %s", saved.Code, saved.Body.String())
	}
	ok := get(e.h, "/ui/ex/posts", session)
	if ok.Code != http.StatusOK {
		t.Fatalf("with the credential: want 200, got %d", ok.Code)
	}
	contains(t, ok, "Hello")
	if e.posts.cred != "secret" {
		t.Fatalf("credential not sent upstream, saw %q", e.posts.cred)
	}
	// The key is never read back into a page.
	if strings.Contains(get(e.h, "/ui/credentials", session).Body.String(), "secret") {
		t.Error("the credentials page must not show the stored key")
	}

	// Another browser session does not share it.
	contains(t, get(e.h, "/ui/ex/posts"), "needs a credential")

	cleared := post(e.h, "/ui/credentials", url.Values{"_leaf": {"p.0"}, "_act": {"0"}}, session)
	if cleared.Code != http.StatusSeeOther {
		t.Fatalf("clearing: want 303, got %d", cleared.Code)
	}
	contains(t, get(e.h, "/ui/ex/posts", session), "needs a credential")
}

func TestAWrongCredentialIsSaidToBeRefused(t *testing.T) {
	e := newEnv(t, resource.AuthScheme{Type: "apikey", In: "header", Name: "X-API-Key"})
	e.posts.needCred = "secret"
	e.build(t)
	first := get(e.h, "/ui/ex/posts")
	session := first.Result().Cookies()[0]
	post(e.h, "/ui/credentials", url.Values{"_leaf": {"p.1"}, "f0": {"wrong"}}, session)
	contains(t, get(e.h, "/ui/ex/posts", session), "refused the credential")
}

func TestBooleansAndNumbersCanBeUnset(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	// featured and rating left empty: both must be omitted, not sent as false / 0.
	rec := post(e.h, "/ui/ex/posts/new", url.Values{"_leaf": {"p"}, "f0": {"Plain"}, "f4": {""}, "f3": {""}})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d: %s", rec.Code, rec.Body.String())
	}
	b := e.posts.lastBody
	if _, has := b["featured"]; has {
		t.Errorf("an unset boolean must be omitted: %#v", b)
	}
	if _, has := b["rating"]; has {
		t.Errorf("an unset number must be omitted: %#v", b)
	}
	// A real 0 is kept.
	post(e.h, "/ui/ex/posts/new", url.Values{"_leaf": {"p"}, "f0": {"Zero"}, "f3": {"0"}})
	if v, has := e.posts.lastBody["rating"]; !has || v != 0.0 {
		t.Errorf("a typed 0 must be kept: %#v", e.posts.lastBody)
	}
}

func TestNoCredentialsPageWithoutAPIKeyBackends(t *testing.T) {
	e := newEnv(t, resourceNone())
	e.build(t)
	if get(e.h, "/ui/credentials").Code == http.StatusOK {
		t.Error("no backend takes a key: there is nothing to configure")
	}
}

func TestRebuildPicksUpNewResourcesAndKeepsSessions(t *testing.T) {
	e := newEnv(t, resourceNone())
	live := mustLive(t, e)
	if get(live, "/ui/ex/tags").Code == http.StatusOK {
		t.Fatal("tags does not exist yet")
	}
	tags := &memStore{rows: []resource.Record{{"id": "t1", "name": "go"}}}
	e.reg.Replace([]*resource.Runtime{
		runtime(t, e.posts, "posts", "Posts", postSchema, resourceNone(), fullCaps),
		runtime(t, tags, "tags", "Tags", authorSchema, resourceNone(), fullCaps),
	})
	if err := live.Rebuild(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	contains(t, get(live, "/ui/ex/tags"), ">go<")
}

func TestUnsafeResourceNamesAreSkippedNotFatal(t *testing.T) {
	e := newEnv(t, resourceNone())
	bad := runtime(t, e.posts, "po{sts", "Bad", postSchema, resourceNone(), fullCaps)
	e.reg.Replace([]*resource.Runtime{bad, runtime(t, e.authors, "authors", "Authors", authorSchema, resourceNone(), fullCaps)})
	e.build(t)
	contains(t, get(e.h, "/ui/ex/authors"), "Ada")
}

func TestRebuildsCanShareOneMetricsRegistry(t *testing.T) {
	e := newEnv(t, resourceNone())
	reg, err := metrics.NewRegistry(metrics.WithNamespace("t"), metrics.WithConstLabels(map[string]string{"env": "test"}))
	if err != nil {
		t.Fatal(err)
	}
	live, err := ui.NewLive(e.reg, ui.Options{PageSize: 25, Metrics: reg.Registerer()})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	get(live, "/ui/ex/posts")
	if err := live.Rebuild(); err != nil {
		t.Fatalf("a second build must not clash with the first one's metrics: %v", err)
	}
	get(live, "/ui/ex/posts")
	rec := get(reg.Handler(), "/metrics")
	contains(t, rec, "webui_http_requests_total", `env="test"`)
}

const personSchema = `{
  "type": "object",
  "required": ["name"],
  "properties": {
    "id": { "type": "string", "readOnly": true },
    "name": { "type": "string", "description": "Shown on the byline." },
    "email": { "type": "string", "format": "email" },
    "site": { "type": "string", "format": "uri" },
    "born": { "type": "string", "format": "date" }
  }
}`

func TestSchemaFormatsAndDescriptionsChooseTheInput(t *testing.T) {
	e := newEnv(t, resourceNone())
	people := runtime(t, &memStore{}, "people", "People", personSchema, resourceNone(), fullCaps)
	e.reg.Replace([]*resource.Runtime{people})
	e.build(t)
	rec := get(e.h, "/ui/ex/people/new")
	contains(t, rec,
		"Shown on the byline.", // description as the field's help
		`type="email"`,         // format: email
		`type="url"`,           // format: uri
		`type="date"`,          // format: date, with no time of day
	)
	if strings.Contains(rec.Body.String(), "datetime-local") {
		t.Error("a date field must not ask for a time of day")
	}
	created := post(e.h, "/ui/ex/people/new", url.Values{"_leaf": {"p"}, "f0": {"Ada"}, "f3": {"1815-12-10"}})
	if created.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d: %s", created.Code, created.Body.String())
	}
}
