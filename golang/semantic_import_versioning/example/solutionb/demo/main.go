package main

import "github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutionb/libfoo"
import libfooV2 "github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutionb/libfoo/v2"

func main(){
	libFooV1 := libfoo.NewClient()
	libFooV2 := libfooV2.NewClient()

	libFooV1.Method4()
	libFooV2.Method4()
}
