package client

import (
	"rsc.io/quote"
	"fmt"
)

func NewClient() {
	fmt.Println("v1 " + quote.Hello())
}
