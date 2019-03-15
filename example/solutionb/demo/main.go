package main

import "github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutionb/libfoo"

func main(){
	libFooV1 := libfoo.NewClient()

	libFooV1.Method4()
}
