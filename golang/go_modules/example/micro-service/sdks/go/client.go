package module

import (
	"fmt"
	"golang.org/x/net/context"

)

type clientImpl struct {
}

func NewClient() Client {
	return &clientImpl{}
}

func (c *clientImpl) API1(){
	// a bug gets fixed
	// another bug gets fixed as well
	fmt.Println("call API1 from the server.")
}

func (c *clientImpl) API2(_ context.Context){
	fmt.Println("call API2 from the server.")
}
