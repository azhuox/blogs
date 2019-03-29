# How to Convert Go Packages to Go Modules

What is included in this blog:
- A brief introduction of Go modules and Semantic Import Versioning
- A discussion about how to convert Go libraries to Go modules
- A discussion about how to convert Go micro services to Go Modules

## prerequisites

### Go Modules

[Go Modules](https://blog.golang.org/modules2019) is an experimental opt-in feature in Go 1.11 with the plan of finalizing feature for Go 1.13. The definition of a Go module from [this proposal](https://go.googlesource.com/proposal/+/master/design/24301-versioned-go.md) is "a group of packages that share a common prefix, the module path, and are versioned together as a single unit". **A file called `go.mod` glues all the files and packages under the same root directory together as a unit.** Here is an example:

```go
path/to/my-repo:
    bar:
        go.mod
        bar-file1.go
        bar-file2.go
        foo:
            foo-file1.go
            foo-file2.go
    mixi:
        go.mod
        mixi-file1.go
        mixi-file2.go
```

[image]

As shown in the picture above, it has two modules `bar` and `mixi`. The module `bar` includes two packages: the package `bar` and the package `foo`.

**The `go.mod` under the `path/to/my-repo/bar` directory glues these two packages together as a single unit.** It defines the module's path and its dependencies:

```go
module path/to/my-repo/bar

require (
	golang.org/x/text v0.3.0
	rsc.io/sampler v1.99.99
	// Other dependencies
)
```

Now the package `bar` and `foo` are bundled together as a unit. For example, the import statement imports the module `path/to/my-repo/bar` other than the package `path/to/my-repo/bar/foo` if Go Modules is enabled. **This means the path in the import statement is considered the module path, not the import path (package path).**

```go
import "path/to/my-repo/bar/foo"

func main () {
    foo.DoSomething()
}
```

In order to use Go Modules, you need to upgrade your Go to v1.11 and set the environmental variable `GO111MODULE=on`.


### Semantic Import Versioning

[Semantic Import Versioning](https://research.swtch.com/vgo-import) is a method proposed for adopting [Semantic Versioning](https://semver.org/) in Go packages or modules. The idea behind it is to embedding major version (say `v2`) in the import path (for packages) or the module path (for modules) with the following rules:

- `v1` must be omitted from the import path or the module path. [This post](https://github.com/golang/go/issues/24301#issuecomment-371228664) explains the reason.
- Major versions higher than `v1` must be embedded in the import path or the module path so that Semantic Versioning can be adopted in Go packages or Go Modules.

The following picture demonstrates the rules above:

[image]

#### Releasing

With Go Modules and Semantic Import Versioning, you can release your modules by creating git tags, for example:

```go
git tag bar/v2.3.3 && git push -q origin master bar/v2.3.3
```

**The tag MUST follow the format {pure_module_path}/v{Major}.{Minor}.{Patch} and {pure_module_path} means the module path without repo URL (which is `bar` in this case). The is key point that makes Go able to retrieve Go modules.**

I recommend you reading [this proposal]((https://research.swtch.com/vgo-import)) or [this blog]() if you want to know more details about Semantic Import Versioning.


All in all, Go Modules provides a way to group one or more packages as a single retrievable unit, while Semantic Import Versioning proposes a method for adopting Semantic Import in Go packages and go Modules. There two things are the foundation of "versioned Go modules".


# What Is A Go Module

Summary:
- Group one or more packages together as a single unit
- Apply Semantic Versioning to these modules so that they can be released/retrieved by Go.


# How to Enable Go Modules

# How to Convert Go Packages to Go Modules

## General Guide

### Converting One or More Packages to A Module

### Releasing


## Converting Go Libraries

## Converting Go Micro Services

