package module

import (
	"fmt"
	"golang.org/x/net/context"

	pkga "github.com/azhuox/blogs/golang/go_modules/example/micro-service/internal/pkga"
	pkgb "github.com/azhuox/blogs/golang/go_modules/example/micro-service/internal/pkgb"
)

type clientImpl struct {
	pkgAClient pkga.Client
	pkgBClient pkgb.Client
}

func NewClient(pkgAClient pkga.Client, pkgBClient pkgb.Client) Client {
	return &clientImpl{
		pkgAClient: pkgAClient,
		pkgBClient: pkgBClient,
	}
}

func (c *clientImpl) API1(){
	fmt.Println("API 1, calling Method1 in pkga")
	c.pkgAClient.Method1()
}

func (c *clientImpl) API2(ctx context.Context){
	fmt.Println("API 2, calling Method2 in pkgb")
	c.pkgBClient.Method2(ctx)
}
