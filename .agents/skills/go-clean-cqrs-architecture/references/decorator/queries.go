package decorator

import "context"

type QueryHandler[T any, E any] interface {
	Handle(ctx context.Context, req T) (E, error)
}

func ApplyQueryDecorator[T any, E any](base QueryHandler[T, E]) QueryHandler[T, E] {
	return &queryDecorator[T, E]{base: base}
}

type queryDecorator[T any, E any] struct {
	base QueryHandler[T, E]
}

func (d *queryDecorator[T, E]) Handle(ctx context.Context, req T) (E, error) {
	// Pre-processing logic can be added here,
	// For example, logging, validation, etc.
	result, err := d.base.Handle(ctx, req)
	if err != nil {
		// Handle error if needed
		return result, err
	}

	// Post-processing logic can be added here
	return result, nil
}
