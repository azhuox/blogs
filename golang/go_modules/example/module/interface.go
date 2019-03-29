package module

import "golang.org/x/net/context"

// Client defines public interface provided by this lib
//
type Client interface {
	Method1()
	Method2(ctx context.Context)
	Method3(ctx context.Context)
}
