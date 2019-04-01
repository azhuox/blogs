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

func (c *clientImpl) Method1(){
	fmt.Println("Method1 in this pkg B")
}

func (c *clientImpl) Method2(_ context.Context){
	fmt.Println("Method2 in pkg B")
}
