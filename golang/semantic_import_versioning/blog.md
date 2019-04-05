# Semantic Import Versioning

What is included in this blog:

- A brief introduction of Go Modules.
- A brief introduction of Semantic Import Versioning.
- A discussion about how to adopt Semantic Import Versioning in Go packages and modules.


## prerequisites

### Semantic Versioning

[Semantic Versioning](https://semver.org/) (semver) is currently the most widely used version scheme in [Software Versioning](https://en.wikipedia.org/wiki/Software_versioning). It uses a sequence of three digital numbers with the format `Major.Minor.Patch` and follows the following rules to indicate a unique status of computer software:

- Increase the `Major` version when you make incompatible breaking changes to the software.
- Increase the `Minor` version when you add backward-compatible features to the software.
- Increase the `Patch` version when you fix bugs in a backward-compatible manner for the software.

### Go Modules

[Go Modules](https://blog.golang.org/modules2019) is an experimental opt-in feature in Go 1.11 with the plan of finalizing feature for Go 1.13. The definition of a Go module from [this proposal](https://go.googlesource.com/proposal/+/master/design/24301-versioned-go.md) is "a group of packages that share a common prefix, the module path, and are versioned together as a single unit". In my opinion, the idea behind Go modules is to break a giant Go repository, for example, a repository with multiple Go libraries, into multiple modules and apply Semantic Versioning to these modules to solve [dependency hell](https://en.wikipedia.org/wiki/Dependency_hell) problems, like conflicting dependencies and diamond dependency.

Here is an example of Go Modules:

```go
my-repo:
    my-thing:
        go.mod
        my-pkg-1:
            file1.go
            file2.go
        my-pkg-2:
            file1.go
            file2.go
```

The `go.mod` file in the `my-thing` folder:

```go
module github.com/path/to/my-thing

go 1.12

require (
 golang.org/x/net v0.0.0-20190313220215-9f648a60d977
 ...
)

```

The `go.mod` file defines the module's path and its dependencies. In this example, the module path in `go.mod` indicates it glues anything under the `my-thing` folder (the `my-pkg-1` package and the `my-pkg-2` package) as a module, which means `my-pkg-1` and `my-pkg-2` now are released and retrieved together as a single unit.

Here is the summary of the relationship between a package, a module, and a repository.

- A package is essentially a directory with some code files. It provides code reusability across Go applications.
- A module consists of one or more packages. It groups these packages as a unit, which is released and retrieved together by Go (after 1.11).
- A repository normally contains a group of Go modules and Go packages.


## Semantic Import Versioning

[Semantic Import Versioning](https://research.swtch.com/vgo-import) is a package management method designed for adopting Semantic Versioning in Go packages and modules. It follows two rules:

- **The import compatibility rule: If an old package (say `libfoo`) and a new package have the same import path, the new package must be backward-compatible with the old one.**
- **If a breaking change occurs, a new package with a different import path (say `libfoo/v2`) must be introduced to distinguish it from the old one.**

## What Kinds of Problems It Can Solve?

### Conflicting Dependencies
The following picture shows the scenario of Conflicting Dependencies, where application A depends on `libfoo` `v1.2.0` and one of its dependencies `libb` requires `libfoo` `v1.9.0`. But different versions of `libfoo` cannot be simultaneously installed. Semantic Import Versioning solves this problem using [Minimal Version Selection Algorithm](https://research.swtch.com/vgo-mvs): The version selected by minimal version selection is always the semantically highest of the versions. In this case, `libfoo` `v1.9.0` is selected as it is the highest version. Moreover, based on the import compatibility rule, `v1.9.0` should be backward-compatible with `v1.2.0` as they have the same import path (which means they have the same `Major` version). Therefore, the application should be able to work with `v1.9.0` without any problem even though it requires `v1.2.0`.

![Conflicting Dependencies](https://raw.githubusercontent.com/aaronzhuo1990/blogs/master/golang/semantic_import_versioning/go-semver-conflicting-dependencies.png)


### Diamond Dependency

The following picture shows the scenario of Diamond Dependency, where application A depends on `libb` and `libc`. Both of them depend on `libd`, but `libb` requires `libd` `v1.1.0` and `libc` requires `libd` `v2.2.2`. Semantic Import Versioning solves this problem by installing both versions and distinguishing them with import paths (`path/to/libd` v.s. `path/to/libd/v2`). `libd` `v1` and `v2` are considered two different packages as they have different import paths.

![Diamond Dependency](https://raw.githubusercontent.com/aaronzhuo1990/blogs/master/golang/semantic_import_versioning/go-semver-diamond-denpendency.png)

## Example

I wrote a dummy package called `libfoo` to demonstrate how Semantic Import Versioning works. You can check [this repo](https://github.com/aaronzhuo1990/blogs/tree/master/golang/semantic_import_versioning/example/) for more details about this example.

Let us go through this example to see how Semantic Import Versioning works.


### Change Log

A file called `CHANGELOG.md` (under the root folder of the package) is used to record the release history of the package. Suppose the following versions need to be released:

```go
--------------------------------------------------------------------------------
## v2

### 2.1.0
- Add `Method6`
- Date: 2019-02-14

### 2.0.0
- BREAKING CHANGE: Modify the signature of the method `Method5` to let it accept an integer and a string
- Date: 2019-02-13

--------------------------------------------------------------------------------
## v1

### 1.1.1
- Fix a bug in `Method4`
- Date: 2019-02-15

### 1.1.0
- Add `Method5`
- Date: 2019-02-12

### 1.0.0
- Production-ready release
- Date: 2019-02-11

--------------------------------------------------------------------------------
## v0

### 0.4.0
- Add `Method4`
- Date: 2019-02-10

### 0.2.2
- Fix a bug in `Method1`
- Date: 2019-02-09

### 0.2.1
- Fix a bug in `Method2`
- Date: 2019-02-08

### 0.2.0
- Add `Method2`
- Add `Method3`
- Date: 2019-02-07

### 0.1.0
- Initial Release
- Add `Method1`
- Date: 2019-02-06
```

From the changelog, you can see that:

- The initial development release starts at `0.1.0` and the `Minor` and `Patch` version is increased respectively for each subsequent release and bug fix.
- `1.0.0` is released when the package is ready for production. There is no breaking change between `v0` and `v1`. `v0` is for internal development while `v1` means most of the bugs are fixed, all the features are fully tested and it can be used in production with the stability guarantee.
- `v2` comes out when a breaking change (modified the signature of Method5()) is made, which means `v2` is incompatible with `v1`.
- `v0` stops releasing any subversion when `v1` comes out, but `v1` and `v2` can be developed individually. For example, you can see that `v1.1.1` is released after `v2.1.0`.
- These releases strictly follow [Semantic Versioning Specification](https://semver.org/spec/v2.0.0.html#semantic-versioning-specification-semvers)

### Problem
**The major problem here is how to how to release `v2` in Go. This is what Semantic Import Versioning is trying to solve.** There are two methods to realize Semantic Import Versioning: Major subdirectory and Major Branch.


## Method A: Major subdirectory

This method actually separates `v1` and `v2` into two packages by giving each of them its own root directory. Here is how the `libfoo` package is organized in this solution:

```go
libfoo/
|-- CHANGELOG.md
|-- client.go
|-- interface.go
|-- v2/
  |-- client.go
  |-- interface.go
```

You can see that `v1` and `v2` are essentially two packages as each of them has its own root directory and import path (`github.com/path/to/libfoo` v.s. `github.com/path/to/libfoo/v2`). The initial codebase of `v2` is copied from `v1`.  **The idea behind this solution is hard-coding `v2` in the import path (by making `v2` subdirectory) to indicate the package's `Major` version.** The following picture demonstrates this idea:

![v1 v.s. v2](https://raw.githubusercontent.com/aaronzhuo1990/blogs/master/golang/semantic_import_versioning/go-semver-v1-vs-v2.png)

From the picture you can see that:

- **`v1` (and `v0`) is omitted from import paths and this is mandatory in go modules. Therefore, you'd better follow this principle if you are thinking of converting your packages into go modules one day.** You can check [this discussion](https://github.com/golang/go/issues/24301#issuecomment-371228664) if you are curious about why they made such a requirement.
- `v2` in the import path indicates the package's `Major` version.
- A single build can use both `v1` and `v2` as they are essentially two packages.
- It does not require to convert your packages to go modules.

### Make It Work with Go Modules

It is very easy to convert both `v1` and `v2` to Go modules. What we need to do is run the following commands to convert them to Go modules:

```go
cd /path/to/solutiona/libfoo
go mod init github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo
go: creating new go.mod: module github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo
go build

cd /path/to/solutiona/libfoo/v2
go mod init github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo/v2
go: creating new go.mod: module github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo/v2
go build
```

It respectively creates a `go.mod` file for `v1` and `v2`:

`v1's` `go.mod`:

```go
module github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo

require rsc.io/quote v1.5.2
```

`v2's` `go.mod`:

```go
module github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo/v2

require rsc.io/quote v1.5.2
```

Take `v1's` `go.mod` as an example, It declares `libfoo` (`v1`) as a module and then lists all of its dependencies. `v1` and `v2` are considered two different go modules as they own different module paths.


### Releasing

#### Without Go Modules

Without Go Modules, you can release the versions listed in the `CHANGELOG.md` file by either creating git tags or creating [Github Releases](https://help.github.com/en/articles/creating-releases) (Creating a GitHub release is essentially creating a git tag). However, without Go Modules enabled, you will not be able to install specific versions of `v1` and `v2` simultaneously in a single build. This is because creating a tag or a release is like creating a snapshot for the whole repository, not just for the single package. A `v2` release will also include the latest version of `v1` and vice versa. **Moreover, the existing Go package management tool can only retrieve dependencies with the repository granularity, not the package or module granularity.** Suppose `libfoo` has released the following versions (with the order from up to down) and you want to use `v1.0.0` and `v2.1.0` in a single build:

```
v1.0.0
v1.1.0
v2.0.0
v2.1.0
v1.1.1
```

You cannot grab `v1.0.0` as it does not have `v2.1.0`. You can only require either `v2.1.0` or `v1.1.1` as they all container `v2.1.0` and `v1.1.x` newer than `1.0.0`. Suppose you require `v2.1.0`, then you will get `v1.1.0` and `v2.1.0` (`v2.1.0` includes all the changes introduced by `v1.1.0`). **In this case, the requirement of `v1.0.0` will not be satisfied but this is OK, as either `v1.1.0` or `v1.1.1` should be backward-compatible with `v1.0.0` provided all the releases follow the import compatibility rule.**


#### With Go Modules

Things become easier when Go Modules is used. With Go Modules, these versions can be released by tagging specific git commits or creating github releases. Here are what I did to release these versions:

1. Cd to the root directory of [example/solutiona/libfoo](https://github.com/aaronzhuo1990/blogs/tree/master/golang/semantic_import_versioning/example/solutiona/libfoo).
2. Realize `Method1() - Method4()` in v1 and comit/push changes: `git commit -q  -m "Realize Method1() - Method4() to release v1.0.0" && git push origin master -q`
3. Create a tag for the changes: `git tag golang/semantic_import_versioning/example/solutiona/libfoo/v1.0.0 && git push -q origin master golang/semantic_import_versioning/example/solutiona/libfoo/v1.0.0`
4. Realize `Method5()` in v1 interface, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutiona/libfoo/v1.1.0`
5. Duplicate `v1` code in the [example/solutiona/libfoo/v2](https://github.com/aaronzhuo1990/blogs/tree/master/golang/semantic_import_versioning/example/solutiona/libfoo/v2) folder, modify the signature of Method5(), commit/push changes and create the tag `golang/semantic_import_versioning/example/solutiona/libfoo/v2.0.0`
6. Realize `Method6()` in v2 interface, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutiona/libfoo/v2.1.0`
7. Pretend to fix a bug in `Method4()` in `v1`, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutiona/libfoo/v1.1.1`
8. Cd to [example/solutiona/demo](https://github.com/aaronzhuo1990/blogs/tree/master/golang/semantic_import_versioning/example/solutiona/demo), create `main.go` and add the following code:

```go
package main

import "github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo"
import libfooV2 "github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo/v2"

func main(){
 libFooV1 := libfoo.NewClient()
 libFooV2 := libfooV2.NewClient()

 libFooV1.Method4()
 libFooV2.Method4()
}
```

9. Initialize `path/to/solutiona/demo` as a go module: `go mod init github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/demo`

10. Build: `go build`

11. Downgrade `libfoo` `v1` to `v1.0.0`: `go get github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo@v1.0.0`

12. Build again: `go build`

13. Run demo: ./demo

```go
v1 Hello, world.
v2 Hello, world.
```

Key points:

1. **A version is released by creating a tag and the tag MUST follow the format {module_path}/v{Major}.{Minor}.{Patch}. This is the key point to make the module retrievable as a single unit by Go.** Take the tag `golang/semantic_import_versioning/example/solutiona/libfoo/v1.1.1` as an example, `golang/semantic_import_versioning/example/solutiona/libfoo` is the module path while `v1.1.1` is the version number. You can see that the repository URL (`github.com/aaronzhuo1990/blogs`) is omitted from the module path.
2. A tag can be created by using `git tag` command or creating a github release, as long as you use the correct format for the tag name. For example, `v1.1.0` is created by using `git tag` command while `v1.1.1` is created by creating a GitHub release. You can check the [release history](https://github.com/aaronzhuo1990/blogs/tags) of this example for more details.
3. `go mod init` always grabs the latest versions of the module's dependencies, which is `libfoo v1.1.1` and `libfoo v2.1.0` in this case. You need to manually downgrade `v1` from `v1.1.1` to `v1.0.0` by running the command `go get github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo@v1.0.0`.
4. A tag or a release is essentially a snapshot for the whole repo, which is `github.com/aaronzhuo1990/blogs` in this example. However, with Go modules, the specific version of `v1` and `v2` can be simultaneously installed. In this case, `v1.0.0` and `v2.1.0` are simultaneously installed in a single build. You can prove this by adding `libFooV1.Method5()` in the demo and run `go build`. It will fail as `v1.0.0` does not have `Method5()`.
5. Please note that in this example, I directly committed/pushed changes into the master branch just for simplifying the demo workflow. This is not a good practice. You are supposed to create a branch, commit/push changes, create a pull request and merge the changes into master branch in real development.


### Advantage of This Solution

- This solution does not require Go Modules even though it has some limitations without Go modules. This means it does not require you to update Go to v1.11 or later versions in order to realize Semantic Import Versioning in your Go packages.
- Its code organization is clear and straightforward as each `Major` version owns its codebase. This allows you to develop `v1` and `v2` very easily.
- It works well with Go modules. This solution allows you to convert your packages to go modules without any problem.


### Disadvantage of This Solution

- Its file structure is somehow strange. From the example, you can see that the root directory of `v1` is `path/to/example/solutiona/libfoo` while the root directory of `v2` is `path/to/example/solutiona/libfoo/v2`. This indicates `v2` lives inside `v1`. The position of `CHANGELOG.md` file also demonstrates this awkwardness.
- A lot of code is duplicated between `v1` and `v2`, as `v2` is generated by duplicating the codebase of `v1`.



## Method B: Major Branch

An alternative way to realize Semantic Import Versioning is to give each `Major` version its own `master` branch. The following steps demonstrate this solution:

1. Cd to the root directory of [example/solutionb/libfoo](https://github.com/aaronzhuo1990/blogs/tree/master/golang/semantic_import_versioning/example/solutionb/libfoo).
2. Realize `Method1() - Method4()` in v1, comit/push changes and create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v1.0.0`
3. Realize `Method5()` in `v1`, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v1.1.0`
4. **Create a branch (say `go-semver-solutionb-libfoo-v1`) based on the master branch for `v1`. We are going to use this branch other than master branch to add features or fix bugs for `v1`.**
5. **Switch back to the master branch, update the go.mod file to include `/v2` at the end of the module path in the module directive (`module golang/semantic_import_versioning/example/solutionb/libfoo/v2`)**
6. Modify the signature of Method5(), commit/push changes and create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v2.0.0`
7. Add Method6(), commit/push changes and create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v2.1.0`
8. Switch back to the branch `go-semver-solutionb-libfoo-v1`, fix a bug in Method4(), commite/push changes and then create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v1.1.1` based on this branch.
9. Create [a demo](https://github.com/aaronzhuo1990/blogs/tree/master/golang/semantic_import_versioning/example/solutionb/demo) to use `v1` and `v2`:

```go
package main

import "github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutionb/libfoo"
import libfooV2 "github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutionb/libfoo/v2"

func main(){
  libFooV1 := libfoo.NewClient()
  libFooV2 := libfooV2.NewClient()

  libFooV1.Method4()
  libFooV2.Method4()
}
```


### Advantage of This Solution

- It solves the code duplication problem between `v1` and `v2`.
- The file structure of `v1` and `v2` makes way more sense.

### Disadvantage of This Solution

- I don't know how to manage `CHANGELOG.md` file for `v1` and `v2`. It looks like each `Major` version needs a `CHANGELOG.md` to record its release history, which means the whole release history of a module is going to be separated into multiple files.
- A repository may get exploded with a bunch of branches when it is managing a lot of modules. Moreover, the master branch is not unique anymore, as an old `Major` version now is using its own branch as its master branch.
- This solution only works with Go modules.

## Summary

- Go modules provide a way for you to group one or more packages into a single unit, while Semantic Import Versioning is a method for adopting Semantic Versioning into Go packages and modules.
- There are two ways to realize Semantic Import Versioning and each of them has its own advantage and disadvantage: The first method (Major Subdirectory) is more straightforward. It can work without Go Modules and allows you to convert your packages to modules very easily. However, it duplicates a lot of code. The second method (Major Branch) does not duplicate any code but may explode a repository since each old `Major` version needs its own `master` branch. Additionally, it only works with Go Modules.
- It is mandatory **NOT** to put `v1` into package path or module path in Go Modules. So you may want to stop using `v1` as a subdirectory if you are thinking of converting your packages to go modules one day.

## What Is Next?

This blog is majorly talking about how to realize Semantic Import Versioning other than Go modules. You may want to read [this article](https://github.com/golang/go/wiki/Modules) if you are curious about go modules.


Reference:

- [Semantic Versioning](https://semver.org/)
- [Software Versioning](https://en.wikipedia.org/wiki/Software_versioning)
- [Go Modules](https://blog.golang.org/modules2019)
- [Proposal: Versioned Go Modules](https://go.googlesource.com/proposal/+/master/design/24301-versioned-go.md)
- [Dependency Hell](https://en.wikipedia.org/wiki/Dependency_hell)
- [Semantic Import Versioning](https://research.swtch.com/vgo-import)
- [Minimal Version Selection](https://research.swtch.com/vgo-mvs)
- [An Example of Semantic Import Versioning](https://github.com/aaronzhuo1990/blogs/tree/master/golang/semantic_import_versioning/example/)
- [Semantic Versioning Specification](https://semver.org/spec/v2.0.0.html#semantic-versioning-specification-semvers)
- [Creating Github Releases](https://help.github.com/en/articles/creating-releases)

