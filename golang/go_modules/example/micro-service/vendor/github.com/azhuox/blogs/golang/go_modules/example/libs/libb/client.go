package liba

import (
	"fmt"
	"golang.org/x/net/context"
	"github.com/azhuox/blogs/golang/go_modules/example/libs/liba"
)

type clientImpl struct {
	libAClient liba.Client
}

func NewClient() Client {
	return &clientImpl{
		libAClient: liba.NewClient(),
	}
}

func (c *clientImpl) Method1(){
	fmt.Println("Method1 in libb, calling liba.Method1():")
	c.libAClient.Method1()
}

func (c *clientImpl) Method2(){
	fmt.Println("Method2 in libb, calling liba.Method2():")
	c.libAClient.Method2(context.TODO())
}
