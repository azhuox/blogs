package main

import (
	"time"
	"golang.org/x/net/context"

	pkga "github.com/azhuox/blogs/golang/go_modules/example/micro-service/internal/pkga"
	pkgb "github.com/azhuox/blogs/golang/go_modules/example/micro-service/internal/pkgb"
	api "github.com/azhuox/blogs/golang/go_modules/example/micro-service/internal/api"
)

func main() {

	pkgAClient := pkga.NewClient()
	pkgBClient := pkgb.NewClient()
	apiClient := api.NewClient(pkgAClient, pkgBClient)

	apiClient.API1()
	apiClient.API2(context.TODO())

	for ; ; {
		time.Sleep(30 * time.Second)
	}
}
