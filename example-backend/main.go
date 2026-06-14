// Command example-backend is a reference "compliant backend" for simple-admin.
//
// It is built on PocketBase (used as a Go framework) and demonstrates the
// contract the admin UI targets: an OpenAPI 3 spec endpoint plus CRUD REST
// endpoints per resource. It is intentionally a SEPARATE Go module so that
// PocketBase's large dependency tree never contaminates the admin UI module.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Olian04/simple-admin-example-backend/openapi"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	// Seed the sample collections once the app is bootstrapped.
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return ensureSampleCollections(e.App)
	})

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Optional static API-key gate on the data API, used to exercise the
		// admin UI's API-key auth passthrough. Enabled by setting EXAMPLE_API_KEY.
		// The OpenAPI spec advertises the matching apiKey security scheme.
		if apiKey := os.Getenv("EXAMPLE_API_KEY"); apiKey != "" {
			e.Router.BindFunc(func(re *core.RequestEvent) error {
				if strings.HasPrefix(re.Request.URL.Path, "/api/collections/") &&
					re.Request.Header.Get("X-API-Key") != apiKey {
					return re.UnauthorizedError("invalid or missing X-API-Key", nil)
				}
				return re.Next()
			})
		}

		// Generated OpenAPI 3 spec describing the collections as CRUD resources.
		e.Router.GET("/openapi.json", func(re *core.RequestEvent) error {
			spec, err := openapi.Generate(re.App, os.Getenv("EXAMPLE_API_KEY") != "")
			if err != nil {
				return re.InternalServerError("failed to generate openapi spec", err)
			}
			return re.JSON(http.StatusOK, spec)
		})

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// ensureSampleCollections creates the demo "authors" and "posts" collections
// (idempotently). Access rules are public ("") so the admin UI can read/write
// in local development without PocketBase auth; the optional X-API-Key gate
// above layers on top when configured.
func ensureSampleCollections(app core.App) error {
	public := ""

	authors, err := app.FindCollectionByNameOrId("authors")
	if err != nil || authors == nil {
		authors = core.NewBaseCollection("authors")
		authors.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 100},
			&core.EmailField{Name: "email"},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		setPublicRules(authors, &public)
		if err := app.Save(authors); err != nil {
			return err
		}
	}

	if posts, err := app.FindCollectionByNameOrId("posts"); err != nil || posts == nil {
		posts = core.NewBaseCollection("posts")
		posts.Fields.Add(
			&core.TextField{Name: "title", Required: true, Max: 200},
			&core.EditorField{Name: "body"},
			&core.SelectField{Name: "status", Values: []string{"draft", "published", "archived"}, MaxSelect: 1},
			&core.NumberField{Name: "rating"},
			&core.BoolField{Name: "featured"},
			&core.RelationField{Name: "author", CollectionId: authors.Id, MaxSelect: 1},
			&core.DateField{Name: "publishedAt"},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		setPublicRules(posts, &public)
		if err := app.Save(posts); err != nil {
			return err
		}
	}

	return nil
}

func setPublicRules(c *core.Collection, rule *string) {
	c.ListRule = rule
	c.ViewRule = rule
	c.CreateRule = rule
	c.UpdateRule = rule
	c.DeleteRule = rule
}
