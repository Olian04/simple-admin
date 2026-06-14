package resource

import "context"

// Service orchestrates CRUD for one resource over its Store + Validator ports.
// Writes are validated before they reach the upstream Store.
type Service struct {
	Descriptor ResourceDescriptor
	Store      Store
	Validator  Validator
}

// Runtime bundles a resource's descriptor with its ready-to-use Service.
type Runtime struct {
	Descriptor ResourceDescriptor
	Service    *Service
}

func (s *Service) List(ctx context.Context, params ListParams) (ListResult, error) {
	return s.Store.List(ctx, params)
}

func (s *Service) Get(ctx context.Context, id string) (Record, error) {
	return s.Store.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, body Record) (Record, error) {
	if err := s.validate(body); err != nil {
		return nil, err
	}
	return s.Store.Create(ctx, body)
}

func (s *Service) Update(ctx context.Context, id string, body Record) (Record, error) {
	if err := s.validate(body); err != nil {
		return nil, err
	}
	return s.Store.Update(ctx, id, body)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.Store.Delete(ctx, id)
}

func (s *Service) validate(body Record) error {
	if s.Validator == nil {
		return nil
	}
	return s.Validator.Validate(body)
}
