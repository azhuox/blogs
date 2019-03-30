package liba

import (
	"fmt"
	"github.com/aaronzhuo1990/blogs/golang/go_modules/example/libs/liba"
	libb "github.com/aaronzhuo1990/blogs/golang/go_modules/example/libs/libb"
)

type clientImpl struct {
	libAClient liba.Client
	libBClient libb.Client
}

func NewClient() Client {
	return &clientImpl{
		libAClient: liba.NewClient(),
		libBClient: libb.NewClient(),
	}
}

func (c *clientImpl) Method1(){
	fmt.Println("Method1 in libc, calling liba.Method1():")
	c.libAClient.Method1()
}

func (c *clientImpl) Method2(){
	fmt.Println("Method2 in libc, calling libb.Method2():")
	c.libBClient.Method2()
}
