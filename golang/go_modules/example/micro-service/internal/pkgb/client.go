package module

import (
	"fmt"
	"golang.org/x/net/context"
	libc "github.com/azhuox/blogs/golang/go_modules/example/libs/libc"
)

type clientImpl struct {
	libcClient libc.Client
}

func NewClient() Client {
	return &clientImpl{
		libcClient: libc.NewClient(),
	}
}

func (c *clientImpl) Method1(){
	fmt.Println("Method1 in pkg B")
}

func (c *clientImpl) Method2(_ context.Context){
	fmt.Println("Method2 in pkg B, calling Method2 in libc")
	c.libcClient.Method2()
}
