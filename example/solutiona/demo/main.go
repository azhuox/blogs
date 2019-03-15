package main

import "github.com/aaronzhuo1990/blogs/example/solutiona/libfoo"
import libfooV2 "github.com/aaronzhuo1990/blogs/example/solutiona/libfoo/v2"

func main(){
	libFooV1 := libfoo.NewClient()
	libFooV2 := libfooV2.NewClient()

	libFooV1.Method4()
	libFooV2.Method4()
}
