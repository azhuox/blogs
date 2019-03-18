package model

import (
	"fmt"
	"golang.org/x/net/context"
)

func NewModel(ctx context.Context) {
	fmt.Println("haha %#v", ctx)
}
