package module

import (
	"fmt"
	"rsc.io/quote"
)

type clientImpl struct {

}

func NewClient() Client {
	return &clientImpl{}
}

func (c *clientImpl) Method1(){
	fmt.Println("Method1 in internal pkg, " + quote.Hello())
}
