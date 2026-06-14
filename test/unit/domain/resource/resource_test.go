package resource_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

func rt(backendID, name string) *resource.Runtime {
	return &resource.Runtime{
		Descriptor: resource.ResourceDescriptor{
			Backend: resource.Backend{ID: backendID, Label: backendID},
			Name:    name,
		},
	}
}

func TestRegistryReplaceAndLookup(t *testing.T) {
	reg := resource.NewRegistry()
	reg.Replace([]*resource.Runtime{
		rt("a", "posts"), rt("a", "users"), rt("b", "items"), nil,
	})

	if _, ok := reg.Lookup("a", "users"); !ok {
		t.Fatal("expected a/users")
	}
	if _, ok := reg.Lookup("a", "missing"); ok {
		t.Fatal("did not expect a/missing")
	}

	backends := reg.Backends()
	if len(backends) != 2 || backends[0].ID != "a" || backends[1].ID != "b" {
		t.Fatalf("backend order wrong: %+v", backends)
	}
	if got := reg.ResourcesFor("a"); len(got) != 2 {
		t.Fatalf("want 2 resources for a, got %d", len(got))
	}
	first, ok := reg.First()
	if !ok || first.Descriptor.Name != "posts" {
		t.Fatalf("First wrong: %+v", first)
	}
}

type fakeStore struct{ created bool }

func (f *fakeStore) List(context.Context, resource.ListParams) (resource.ListResult, error) {
	return resource.ListResult{}, nil
}
func (f *fakeStore) Get(context.Context, string) (resource.Record, error) { return nil, nil }
func (f *fakeStore) Create(context.Context, resource.Record) (resource.Record, error) {
	f.created = true
	return resource.Record{}, nil
}
func (f *fakeStore) Update(context.Context, string, resource.Record) (resource.Record, error) {
	return resource.Record{}, nil
}
func (f *fakeStore) Delete(context.Context, string) error { return nil }

type failValidator struct{}

func (failValidator) Validate(resource.Record) error {
	return &resource.ValidationError{Fields: map[string]string{"title": "required"}}
}

func TestServiceCreateValidatesBeforeStore(t *testing.T) {
	store := &fakeStore{}
	svc := &resource.Service{Store: store, Validator: failValidator{}}

	_, err := svc.Create(context.Background(), resource.Record{})
	var verr *resource.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("want *ValidationError, got %v", err)
	}
	if store.created {
		t.Fatal("upstream Create must not be called when validation fails")
	}
}

func TestServiceCreateNoValidator(t *testing.T) {
	store := &fakeStore{}
	svc := &resource.Service{Store: store}
	if _, err := svc.Create(context.Background(), resource.Record{}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !store.created {
		t.Fatal("expected upstream Create to run when no validator")
	}
}
