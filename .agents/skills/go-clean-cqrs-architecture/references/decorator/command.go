package decorator

import "context"

type CommandHandler[T any] interface {
	Handle(ctx context.Context, req T) error
}

func ApplyCommandDecorator[req any](base CommandHandler[req]) CommandHandler[req] {
	return &commandDecorator[req]{base: base}
}

type commandDecorator[req any] struct {
	base CommandHandler[req]
}

func (d *commandDecorator[T]) Handle(ctx context.Context, req T) error {
	// Pre-processing logic can be added here,
	// For example, logging, validation, etc.
	err := d.base.Handle(ctx, req)
	if err != nil {
		// Handle error if needed
		return err
	}

	// Post-processing logic can be added here
	return nil
}

type CommandReturnHandler[T any, E any] interface {
	Handle(ctx context.Context, req T) (E, error)
}

func ApplyCommandReturnDecorator[T any, E any](base CommandReturnHandler[T, E]) CommandReturnHandler[T, E] {
	return &commandReturnDecorator[T, E]{base: base}
}

type commandReturnDecorator[T any, E any] struct {
	base CommandReturnHandler[T, E]
}

func (d *commandReturnDecorator[T, E]) Handle(ctx context.Context, req T) (E, error) {
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
