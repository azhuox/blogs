package module

import "golang.org/x/net/context"

// Client defines public interface provided by this lib
//
type Client interface {
	API1()
	API2(ctx context.Context)
}
