package libfoo

import (
	"rsc.io/quote"
	"fmt"
)

// clientImpl - An implementation of Client interface
//
type clientImpl struct {

}

// NewClient - return an `Client` interface
//
func NewClient() Client {
	return &clientImpl{}
}

// Method1 - the implementation of `Method1` method defined in the `Client` interface
//
func (c *clientImpl) Method1() error {

	// Concrete implementation details are omitted

	return nil
}

// Method2 - the implementation of `Method2` method defined in the `Client` interface
//
func (c *clientImpl) Method2() error {

	// Concrete implementation details are omitted

	return nil
}

// Method3 - the implementation of `Method3` method defined in the `Client` interface
//
func (c *clientImpl) Method3() error {

	// Concrete implementation details are omitted

	return nil
}

// Method4 - the implementation of `Method4` method defined in the `Client` interface
//
func (c *clientImpl) Method4() error {

	// Add this comment to pretend to fix a bug in this method

	// Concrete implementation details are omitted

	fmt.Println("solution B: v1 " + quote.Hello())
	return nil
}

// Method5 - the implementation of `Method5` method defined in the `Client` interface
//
func (c *clientImpl) Method5(param1 int, param2 string) error {

	// Concrete implementation details are omitted

	return nil
}