package main

import "github.com/azhuox/blogs/golang/semantic_import_versioning/example/solutiona/libfoo"
import libfooV2 "github.com/azhuox/blogs/golang/semantic_import_versioning/example/solutiona/libfoo/v2"

func main(){
	libFooV1 := libfoo.NewClient()
	libFooV2 := libfooV2.NewClient()

	libFooV1.Method4()
	libFooV2.Method4()
}
