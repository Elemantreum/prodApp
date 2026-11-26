package section

import "context"

type scopeKey struct{}

const defaultSectionID = 1

func ID(ctx context.Context) int {
	if id, ok := ctx.Value(scopeKey{}).(int); ok {
		return id
	}
	return defaultSectionID
}

func WithID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, scopeKey{}, id)
}
