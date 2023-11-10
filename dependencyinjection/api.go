package dependencyinjection

import "context"

func Get(ctx context.Context, key string) any {
	ctn, ok := ctx.Value(containerKey).(*container)
	if !ok {
		panic("container does not exist in context")
	}

	return ctn.Get(key)
}
