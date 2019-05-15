What is included in this blog:

- A brief introduction of Go modules and Semantic Import Versioning
- A discussion about how to convert multiple Go libraries in the same  repository to Go modules
- A discussion about how to utilize Go Modules in microservices

# prerequisites

## Go Modules

[Go Modules](https://blog.golang.org/modules2019) is an experimental opt-in feature in Go 1.11 with the plan of finalizing feature for Go 1.13. The definition of a Go module from [this proposal](https://go.googlesource.com/proposal/+/master/design/24301-versioned-go.md) is "a group of packages that share a common prefix, the module path, and are versioned together as a single unit". It is designed for resolving [dependency hell](https://en.wikipedia.org/wiki/Dependency_hell) problems in Go, like conflicting dependencies and diamond dependency.

### An Example

Here is an example of Go Modules:

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

![Modules in my-repo](https://raw.githubusercontent.com/azhuox/blogs/master/golang/go_modules/images/modules-in-my-repo.png)

As shown in the picture, the `my-repo` repository has two modules `bar` and `mixi`. Take the `bar` module as an example, it contains two packages: the `bar` package and the `foo` package. **The `go.mod` file under the `path/to/my-repo/bar` directory defines the module's path and its dependencies:**

```go
module path/to/my-repo/bar

require (
 golang.org/x/text v0.3.0
 rsc.io/sampler v1.99.99
 // Other dependencies
)
```

The `go.mod` file bundles the `bar` package and the `foo` package together as a unit. For example, the import statement in the following code will import the module `path/to/my-repo/bar` (which includes the `foo` package) rather than the package `path/to/my-repo/bar/foo` when Go Modules is enabled. **Even though the code looks the same, the path in the import statement is recognized as the module path, not the package path, once Go Modules is used**

```go
import "path/to/my-repo/bar/foo"

func main () {
  foo.DoSomething()
}
```

### How to Enable Go Modules

In order to use Go Modules, you need to upgrade your Go to v1.11 or any later version and set the environment variable `export GO111MODULE=on`.

### When to Use Go Modules

**The major purpose of Go Modules is to let one or more packages be versioned, released and retrieved together as a single unit. Therefore, the public packages, for example, Go libraries and SDKs, are major targets of Go Modules as they need to be published properly for public use.** You do not need to convert internal packages or any internal-used-only packages within a microservice repository to Go modules. These packages can directly import and use modules once Go Modules feature is enabled, even if they are not converted to modules.

## Semantic Import Versioning

[Semantic Import Versioning](https://github.com/azhuox/blogs/blob/master/golang/semantic_import_versioning/blog.md) is a method proposed for adopting [Semantic Versioning](https://semver.org/) in Go packages and modules. The idea behind it is embedding the major version (say `v2`) in the package path (for packages) or the module path (for modules) with the following rules:

- `v1` must be omitted from the module path. [This post](https://github.com/golang/go/issues/24301#issuecomment-371228664) explains the reason. You may need to follow this rule in your packages if you are thinking of converting your packages to modules one day.
- The Major versions higher than `v1` must be embedded in the package path or the module path so that Semantic Versioning can be applied to Go packages and modules.

The following picture demonstrates the rules above:

![Semantic Import Versioning](https://raw.githubusercontent.com/azhuox/blogs/master/golang/go_modules/images/semantic-import-versioning.png)

### Releasing

**With Go Modules and Semantic Import Versioning, you can release your modules by creating git tags. A tag corresponds to a version.** For example, the following git command releases the `bar` module `v2.3.3`:

```go
git tag bar/v2.3.3 && git push -q origin master bar/v2.3.3
```

You can read  [my last blog](https://github.com/azhuox/blogs/blob/master/golang/semantic_import_versioning/blog.md) for more details about how to releases modules with Semantic Import Versioning.

**All in all, Go Modules provides a way to group one or more packages as a single retrievable unit, while Semantic Import Versioning is a method for applying Semantic Versioning in Go packages and modules to make them versioned. These two things are designed for breaking a repository into multiple retrievable units (modules), so that Go can grab dependencies at the module granularity rather than the repository granularity.**

# Utilizing Go Modules

## General Guide of Converting Go Packages to Go Modules

I wrote [a dummy package](https://github.com/azhuox/blogs/tree/master/golang/go_modules/example/module) called `module` for demonstrating how to convert one or more Go packages to a Go module.

### Converting

It is very easy to convert one or more Go packages to a Go module. Take the `module` package as an example, here are the steps of converting it to a Go module:

1. Cd to the root directory of the `module` package: `cd path/to/module`
2. Convert the package to a module: `go mod init github.com/azhuox/blogs/golang/go_modules/example/module`
3. Compile the module and its dependencies: `go build`
4. Commit the changes automatically generated by Go: `git add ./go.mod ./go.sum && git commit -q  -m "Convert the package to a module" && git push origin master -q`
5. (Optional) you can run `go mod vendor` to reset the module's vendor directory to include all the packages and modules which are required for building and testing all of the module's packages. This is the way to provide dependencies for the older versions of Go that do not fully understand Go modules. Any version of Go >= v1.11 does not need this.


Here are the contents of the `go.mod` file automatically generated by Go. You can see that it defines the module's path, glues anything under the `path/to/example/module` directory as a single unit and lists all of its dependencies.

```go
module github.com/azhuox/blogs/golang/go_modules/example/module

go 1.12

require (
 golang.org/x/net v0.0.0-20190328230028-74de082e2cca
 rsc.io/quote v1.5.2
)
```

Go utilizes the following roles to grab the module's dependencies:

**1. It grabs the latest version for the packages that have been converted to modules. For example, `rsc.io/quote v1.5.2`.**

**2. It grabs the latest commit for the packages that have not been converted to modules with the format `v0.0.0-{date}-{first_12_characters_of_commit_id}`. For example, `golang.org/x/net v0.0.0-20190328230028-74de082e2cca`.**

### Releasing

A module can only be used as a module after it is released. A module is released by creating git tags and each tag corresponds to a version. However, there are two problems we need to solve before releasing a module.

The first problem is how to release `v2` or higher Major versions. Go utilizes two methods, Major Branch and Major Subdirectory, which are provided by [this proposal](https://research.swtch.com/vgo-module#from_repository_to_modules) to solve this problem. [My last blog](https://github.com/azhuox/blogs/blob/master/golang/semantic_import_versioning/blog.md) demonstrates these two methods and compare their advantages and disadvantages. In this blog, Major Subdirectory is used for all the examples as it does not require to duplicate any code.

The second problem is we need to figure out whether to consider the conversion from Go package(s) to a Go module a breaking change or not. If so, we need to upgrade the Major version using [Semantic Versioning](https://semver.org/). If not, we need to decide what versions we need to release. I prefer to just release the latest version of the package(s) listed in the `CHANGELOG.md` file for the following reasons:

- **The conversion from Go package(s) to a Go module is not a breaking change as the package(s) can still work with older versions of Go even if the package(s) are converted to a module. So it does not make sense to upgrade the Major version for this kind of change.**
- **The conversion from Go package(s) to a Go module does not add any new feature or fix any bug. So upgrading the Minor or Patch version in this case does not make sense either.**


Now let us come back to the module example and release its latest version. Here is what I did:

1. Appended `v2` to the end of the module path (module `github.com/azhuox/blogs/golang/go_modules/example/module/v2`) as the latest version of the `module` package is `v2.0.1`.
2. Add a note under the `v2.0.1` release note in the `CHANGELOG.md` file to indicate that the package is converted to a module in and after this version.
3. Release `v2.0.1` by creating a git tag: `git tag golang/go_modules/example/module/v2.0.1 && git push -q origin master golang/go_modules/example/module/v2.0.1`

### Consuming A Module

You can still use this package, without Go Modules enabled, by using some Go dependency management tool (e.g. `dep`) with the following specification. This will grab the whole repository which includes the `module` module for your build.

```go
[[constraint]]
name = "github.com/azhuox/blogs"
branch = "master"
```

With Go Modules, what you need to do is import and use the module in your Go program and run `go build`. It will automatically grab the `golang/go_modules/example/module/v2.0.1` module other than the whole repository for your build.


## Converting Go Libraries to Go Modules

The section above already demonstrates how to convert one or more Go packages to a Go module. This section majorly talks about how to convert all the Go packages (libraries) within the same repository to Go modules.

I wrote three packages `liba` `libb` and `libc` under the `github.com/azhuox/blogs/golang/go_modules/example/libs/` directory for the demo purpose. Among these three packages, the `libb` package depends on the `liba` package while the `libc` package depends on the `libb` and `libc` package.

**A principle that we need to follow in this case is firstly convert the packages that have no dependency on other packages within the same repository, and then convert the packages which dependencies have been converted Go modules.** This indicates that we need to convert the `liba` package first, then the `libb` package and then the `libc` package in this case.

Let us see what will happen if we convert `libc` first:

```go
go mod init github.com/azhuox/blogs/golang/go_modules/example/libs/libc
go: creating new go.mod: module github.com/azhuox/blogs/golang/go_modules/example/libs/libc
go build:

can't load package: package github.com/azhuox/blogs/golang/go_modules/example/libs/libc: unknown import path "github.com/azhuox/blogs/golang/go_modules/example/libs/libc": ambiguous import: found github.com/azhuox/blogs/golang/go_modules/example/libs/libc in multiple modules:
      github.com/azhuox/blogs/golang/go_modules/example/libs/libc (/Users/achuo/go/src/github.com/azhuox/blogs/golang/go_modules/example/libs/libc)
      github.com/azhuox/blogs v0.0.0-20190330175117-09a7dbd4a3ce (/Users/achuo/go/pkg/mod/github.com/azhuox/blogs@v0.0.0-20190330175117-09a7dbd4a3ce/golang/go_modules/example/libs/libc)
```

The cause of this `ambiguous import` problem is Go grabs the whole repository `github.com/azhuox/blogs v0.0.0-20190330175117-09a7dbd4a3ce` to get the `liba` and `libb` package for satisfying the dependencies of the `libc` module. However, `github.com/azhuox/blogs v0.0.0-20190330175117-09a7dbd4a3ce` also includes a copy of the `libc` package, which confuses the Go compiler. To fix this, we need to convert the `liba` and `libb` package to Go modules and release them, so that they can be retrieved and parsed properly as two individual modules by Go.


Now let us convert these three libs in a correct order.

Convert the `liba` package to a module:

```go
cd path/to/libs/liba
go mod init github.com/azhuox/blogs/golang/go_modules/example/libs/liba
  go: creating new go.mod: module github.com/azhuox/blogs/golang/go_modules/example/libs/liba
go build
  go: finding golang.org/x/net/context latest
  go: finding golang.org/x/net latest

# Commit changes
#
git add ./go.mod ./go.sum
git commit ./go.mod ./go.sum -q -m "Convert liba  to a module" && git push origin master -q

# Release the latest version (v1.1.0):
#
git tag golang/go_modules/example/libs/liba/v1.1.0 && git push -q origin master golang/go_modules/example/libs/liba/v1.1.0
```

convert the `libb` package to a module:

```go
go mod init github.com/azhuox/blogs/golang/go_modules/example/libs/libb
  go: creating new go.mod: module github.com/azhuox/blogs/golang/go_modules/example/libs/libb
go build
  go: downloading github.com/azhuox/blogs/golang/go_modules/example/libs/liba v1.1.0
  go: extracting github.com/azhuox/blogs/golang/go_modules/example/libs/liba v1.1.0
  ...

git add ./go.mod ./go.sum
git commit ./go.mod ./go.sum -q -m "Convert libb  to a module" && git push origin master -q
git tag golang/go_modules/example/libs/libb/v1.0.0 && git push -q origin master golang/go_modules/example/libs/libb/v1.0.0
```

Convert the `libc` package to a module:

```go
go mod init github.com/azhuox/blogs/golang/go_modules/example/libs/libc
go build
  go: downloading github.com/azhuox/blogs/golang/go_modules/example/libs/libb v1.0.0
  go: extracting github.com/azhuox/blogs/golang/go_modules/example/libs/libb v1.0.0
  ...

git add ./go.mod ./go.sum
git commit ./go.mod ./go.sum -q -m "Convert libc  to a module" && git push origin master -q
git tag golang/go_modules/example/libs/libc/v1.0.0 && git push -q origin master golang/go_modules/example/libs/libc/v1.0.0
```

You can see the `libc` package is converted to a module correctly and it can retrieve the `liba` and `libb` modules in its build without any problem.


## Go Modules and Microservices

I wrote [a dummy micro-service](https://github.com/azhuox/blogs/tree/master/golang/go_modules/example/micro-service) for demonstrating how to utilize Go Modules in a microservice. Here is its project layout:

```go
github.com/azhuox/blogs/tree/master/golang/go_modules/example/micro-service:
  - sdks
      - go
  - internal
      - api
      - pkga
      - pkgb
  - server
      - main.go
  - vendor
  - Gopkg.toml
  - Gopkg.lock
  - Dockerfile
```

I want to mention that the `internal/pkgb` package is using `libc` package that we just converted to a Go module above. In this case, `libc` is retrieved together with `liba` and `libb` from the `github.com/azhuox/blogs` repository when Go Modules is not enabled. But it is retrieved individually as a single unit when Go Modules is enabled.

From the project layout, you can also see that the microservice is built as a docker image with the following Dockerfile:

```go
FROM golang:1.12-alpine3.9

RUN apk add --update \
  ca-certificates \
  git

COPY . $GOPATH/src/github.com/azhuox/blogs/golang/go_modules/example/micro-service
RUN go build -o /usr/bin/micro-service github.com/azhuox/blogs/golang/go_modules/example/micro-service/server && rm -rf $GOPATH/*

ENTRYPOINT ["/usr/bin/micro-service"]
```

### Converting Public Packages to Go Modules

As mentioned in the [When to Use Go Modules](#when-to-use-go-modules) section, only public packages need to be converted to modules. In this case, the `sdks/go` package is the only package that gets publicly used. Therefore, we only need to convert this package to a module and releases its latest version:

```go
go mod init github.com/azhuox/blogs/golang/go_modules/example/micro-service/sdks/go
go build
  ...
git add ./go.mod ./go.sum
git commit ./go.mod ./go.sum -q -m "Convert micro-service/sdks/go to a module" && git push origin master -q
git tag golang/go_modules/example/micro-service/sdks/go/v1.0.2 && git push -q origin master golang/go_modules/example/micro-service/sdks/go/v1.0.2
```

### Utilizing Go Modules in the Microservice

Go Modules in this case refers to the new Go package management tool called [vgo](https://github.com/golang/go/wiki/vgo) which is integrated in go tools like `go get` and `go mod`. The following steps demonstrate how to use it to manage the dependencies for the microservice:

1. Launch a terminal and then enable Go Modules in the terminal: `export GO111MODULE=on`.
2. Cd the root directory of the microservice.
3. Add a `go.mod` file to the root directory of the microservice: `go mod init github.com/azhuox/blogs/golang/go_modules/example/micro-service`.
4. Run or test the microservice to ensure that everything works fine: `go run ./server/main.go`. This will generate a file called `go.sum` if everything goes well.
5. Remove the files for the old dependency management tool, which is `Gopkg.toml` and `Gopkg.lock` in this case.
5. Commit the changes.

Now we successfully replace the old dependency management tool with Go Modules. However, there are two cases we need to deal with in the Continuous Integration (CI) process: with vendor or without vendor.


#### CI Without Vendor

Without vendors means utilizing Go Modules to dynamically grab dependencies when building docker images during the CI process. In order to do this, we need to do the following steps:

1. Add an environment variable `ENV GO111MODULE=on` in the Dockerfile to enable Go Modules.
2. Remove the `vendor` directory since we don't need it anymore.
3. Commit the changes.

#### CI With Vendor

With vendor means we want to dump all the dependencies into the `vendor` directory and let the CI build the docker image based on the `vendor` directory. The following steps demonstrate how to do it:

1. Dump all the dependencies into the `vendor` directory: `go mod vendor`.
2. Commit the changes.
3. If Go Modules is enabled in the CI tool, add the `-mod=vendor` in the `go build` step in the Dockerfile: `go build -mod=vendor -o /usr/bin/micro-service github.com/azhuox/blogs/golang/go_modules/example/micro-service/server && rm -rf $GOPATH/*`.

#### Update A Dependency in the `vendor` Directory

Suppose we want to build docker images with the `vendor` directory and use the latest version of `libc` (say v1.5.0) in the microservice. The following steps demonstrates the update process:

1. Get the version: `go get github.com/azhuox/blogs/golang/go_modules/example/libs/libc@v1.5.0`.
2. Update the `vendor` directory: `go mod vendor`.

This may not work when the microservice is not using any new feature released after the current version of `libc` (v1.0.0 in this case). To force update it, we need to add a replace statement in the `go.mod` file and then run `go mod vendor`:

```
replace (
    github.com/azhuox/blogs/golang/go_modules/example/libs/libc v1.0.0 github.com/azhuox/blogs/golang/go_modules/example/libs/libc v1.5.0
)
```


# Summary

- Go Modules allows you group one or more packages to a single unit which is released and retrieved together.
- Semantic Import Versioning is a method for applying Semantic Versioning to Go packages and modules to make them versioned.
- Only the publicly-used packages, for example, Go libraries and SDKs, need to convert to Go modules (which produces the modules).
- It is very easy to replace a legacy Go package management tool (e.g. dep) with Go modules (which consumes the modules).

# Reference

- [Go Modules](https://blog.golang.org/modules2019)
- [Proposal: Versioned Go Modules](https://go.googlesource.com/proposal/+/master/design/24301-versioned-go.md)
- [dependency hell](https://en.wikipedia.org/wiki/Dependency_hell)
- [Semantic Import Versioning](https://research.swtch.com/vgo-import)
- [Semantic Versioning](https://semver.org/)
- [Semantic Import Versioning in Go](https://github.com/azhuox/blogs/blob/master/golang/semantic_import_versioning/blog.md)
- [Defining Go Modules](https://research.swtch.com/vgo-module#from_repository_to_modules)
- [vgo](https://github.com/golang/go/wiki/vgo)

