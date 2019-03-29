package module

import (
	"fmt"
	"rsc.io/quote"
	"golang.org/x/net/context"

	internal "github.com/aaronzhuo1990/blogs/golang/go_modules/example/module/internal"
)

type clientImpl struct {
	intHelper internal.Client
}

func NewClient() Client {
	return &clientImpl{
		intHelper: internal.NewClient(),
	}
}

func (c *clientImpl) Method1(){
	fmt.Println("Method1 in this module")
}

func (c *clientImpl) Method2(_ context.Context){
	fmt.Println("Method2 in this module, " + quote.Hello())
}

func (c *clientImpl) Method3(_ context.Context){
	fmt.Println("Method3 in this module, calling internal.Method1()" )

	// Pretend that a bug is fixed

	c.intHelper.Method1()
}