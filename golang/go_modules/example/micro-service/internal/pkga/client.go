package module

import (
	"fmt"
	"rsc.io/quote"
	"golang.org/x/net/context"
)

type clientImpl struct {
}

func NewClient() Client {
	return &clientImpl{
	}
}

func (c *clientImpl) Method1(){
	fmt.Println("Method1 in this pkg A")
}

func (c *clientImpl) Method2(_ context.Context){
	fmt.Println("Method2 in pkg A, " + quote.Hello())
}
