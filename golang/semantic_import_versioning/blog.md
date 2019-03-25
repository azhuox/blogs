# Semantic Import Versioning In Go

What is included in this blog:
- A brief introduction of Go Modules.
- A brief introduction  of Semantic Import Versioning in Go.
- A discussion about how to to realize Semantic Import Versioing with and without Go module.


## prerequisites

### Semantic Versioning

[Semantic Versioning](https://semver.org/) (semver) is currently the most widely used version scheme in [Software Versioning](https://en.wikipedia.org/wiki/Software_versioning). It uses a sequence of three digital numbers with the format `Major.Minor.Patch` and with the following rules to indicate an unique status of a computer software:
- Increase the `Major` version when you make incompatible breaking changes to the software.
- Increase the `Minor` version when you add backwards-compatible features to the software.
- Increase the `Patch` version when you fix bugs in a backwards-compatible manner for the software.

### Go Modules

Go modules are an experimental opt-in feature in Go 1.11 with the plan of finalizing feature for Go 1.13. The definition of a Go module from [this proposal](https://go.googlesource.com/proposal/+/master/design/24301-versioned-go.md) is "a group of packages that share a common prefix, the module path, and are versioned together as a single unit". In my opinion, the idea behind Go Modules is to break a giant Go repo into multiple modules and adopts Semantic Versioning in these modules to solve [dependency hell](https://en.wikipedia.org/wiki/Dependency_hell) problems, like conflicting dependencies or diamond dependency.

Here is an example:

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

`go.mod` file in the `my-thing` folder:

```go
module github.com/path/to/my-thing

go 1.12

require (
	golang.org/x/net v0.0.0-20190313220215-9f648a60d977
	...
)

```

`go.mod` defines a module's path and its dependencies. In this example, the module path in `go.mod` indicates it groups anything under the `my-thing` folder (`my-pkg-1` and `my-pkg-2`) as a module. This means `my-pkg-1` and `my-pkg-2` now should be released together and can be retrieved as a unit by `go get`.

Here is the summary of the relationship between a package, a module and a repo.

- A Package is essentially a directory with some code files. It provides code reusability across the Go applications.
- A Module consists of one or more packages. It groups the package(s) as a unit, which should be released and can be retrieved together by Go (after 1.11).
- A repo is a group of Go modules and Go packages.


## Semantic Import Versioning

[Semantic Import Versioning](https://research.swtch.com/vgo-import) is a package management method designed for adopting Semantic Versioning in Go packages and modules. It follows the following rules:
- **The import compatibility rule: "If an old package (say `libfoo`) and a new package have the same import path, the new package must be backwards compatible with the old one."**
- **If a breaking change occurs, a new package must be released with a different import path (say `libfoo/v2`) to distinguish it from the old one.**

## What Kinds of Problems It Can Solve?

### Conflicting Dependencies
The following picture shows the scenario of Conflicting Dependencies, in which application A depends on `libFoo` v1.2.0 and one of its dependencies `libB` requires `libFoo` v1.9.0. But different versions of libfoo cannot be simultaneously installed. Semantic Import Versioning solves this problem with the [minimal version selection algorithm](https://research.swtch.com/vgo-mvs): The version selected by minimal version selection is always the semantically highest of the versions. In this case, `libFoo` v1.9.0 is selected as it is the highest version. Moreover, based on the specification of Semantic Versioning, v1.9.0 should be back-compatible with v1.2.0 as they have the same `Major` version. Therefore, the application should be able to works with v1.9.0 without any problem even though it requires v1.2.0.
[image]

### Diamond Dependency

The following picture shows the scenario of Diamond Dependency, in which application A depends on `libB` and `libC`, both B and C depends on `libD`, but B requires D v1.1.0 and C requires D v2.2.2. Semantic Import Versioning solves this problem by installing both versions and distinguishing them with import paths, for example, `path/to/d` v.s. `path/to/d/v2`.
[image]


## Example

I wrote a dump package called `libfoo` to demonstrate how Semantic Import Versioning works. You can check [this repo](https://github.com/aaronzhuo1990/blogs/tree/master/golang/semantic_import_versioning/example/) for more details about this example.

Let us go through this example to see how Semantic Import Versioning works.


### Change Log

A file called `CHANGELOG.md` (under the root folder of the package) is used to record release history of the package. Suppose the following versions need to be released:

```
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

From the change log, you can see that:

- The initial development release starts at `0.1.0` and the `Minor` and `Patch` version are increased respectively for each subsequent release and each bug fix release.
- `1.0.0` is released when the package is ready for production. There is no breaking changes between `v0` and `v1`. `v0` is for internal development while `v1` means most of bugs are fixed, all the features are fully tested and it can be used in production with the stability guarantee.
- `v2` comes out as a breaking change is made. `v2` is incompatible with `v1`.
- `v0` stops releasing any sub version when `v1` comes out, but `v1` and `v2` can be developed individually. For example, you can see that `v1.1.1` is released after `v2.1.0`.
- These releases strictly follow [Semantic Versioning Specifications](https://semver.org/spec/v2.0.0.html#semantic-versioning-specification-semvers)

### Problem
**The major problem here is how to how to release `v2` in Go. This is what Semantic Import Versioning is trying to solve.**


## Method A: Major subdirectory

This method actually separates `v1` and `v2` into two packages. Here is the how the `libfoo` package is organized in this solution:

```go
libfoo/
|-- CHNAGELOG.md
|-- client.go
|-- interface.go
|-- v2/
    |-- client.go
    |-- interface.go
```

You can see that `v1` and `v2` are actually two packages as each of them has its own root directory and import path (`github.com/path/to/libfoo` and `github.com/path/to/libfoo/v2`). The initial codebase of `v2` is copied from `v1`.  The idea behind this solution is embedding `v2` in the import path to indicate the package's `Major` version. The following picture shows this relationship:

[image]

From the picture you can see that:

- **`v1` (and `v0`) is omitted from import paths and this is mandatory in go modules. Therefore, you'd better follow this principle if you are thinking of converting your packages into go modules one day.** You can check [this discussion](https://github.com/golang/go/issues/24301#issuecomment-371228664) if you are curious about why they made such requirement.
- `v2` in the import path indicates the package's `Major` version.
- A single build can use both `v1` and `v2` as they are essentially two packages.
- It does not require to covert your packages to go modules.

### Make It Work with Go Modules

It is very easy to convert the package `libfoo` to a go module. What we need to do is run the following commands to convert `v1` and `v2` to go modules:

```
cd /path/to/solutiona/libfoo
go mod init github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo
go: creating new go.mod: module github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo
go build

cd /path/to/solutiona/libfoo/v2
go mod init github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo/v2
go: creating new go.mod: module github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo/v2
go build
```

It then respectively creates a `go.mod` file for `v1` and `v2`:

`v1's` `go.mod`:

```yaml
module github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo

require rsc.io/quote v1.5.2
```

`v2's` `go.mod`:

```yaml
module github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo/v2

require rsc.io/quote v1.5.2
```

Take `v1's` `go.mod` as an example, It declares `libfoo` (`v1`) as a module and then lists all the dependencies. `v1` and `v2` are considered two different go modules.


### Releasing

#### Without Go Modules

Without Go Modules, you can release the versions listed in the `CHANGELOG.md` file by either creating git tags or creating [github releases](https://help.github.com/en/articles/creating-releases). However, this does not allow you install specific versions of `v1` and `v2` simultaneously in a single build. This is because creating a tag or release is like creating a snapshot for the whole repo, not just for the single package. For example, suppose you want to use `v1.0.0` and `v2.1.0` in a single build and `libfoo` has released the following versions:

```
v1.0.0
v1.1.0
v2.0.0
v2.1.0
v1.1.1
```

You cannot grab `v1.0.0` as it does not have `v2.1.0`. You can only require either `v2.1.0` or `v1.1.1` as they all container `v2.1.0` and `v1.1.x` version later than `1.0.0`. Suppose you require `v2.1.0`, then you will get `v1.1.0` and `v2.1.0` (**`v2.1.0` includes all the changes introduced by `v1.1.0`**). **In this case, the requirement of `v1.0.0` will not be satisfied but this is OK as either `1.1.0` or `1.1.1` is supposed to back-compatible with `1.0.0` because of the specification of Semantic Versioning.**


#### With Go Modules

Things become easier when using Go Modules. With Go Modules, these versions can be released by tagging specific commits using git or creating github releases. Here are what I did to release these versions:

1. Cd to the root directory of `solutiona`.
2. Realize `Method1() - Method4()` in v1 and comit/push changes. **Realize Method1() - Method4() to release v1.0.0" && git push origin master -q**
3. Create a tag for the above changes: *git tag golang/semantic_import_versioning/example/solutiona/libfoo/v1.0.0 && git push -q origin master golang/semantic_import_versioning/example/solutiona/libfoo/v1.0.0*
4. Realize `Method5()` in v1 interface, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutiona/libfoo/v1.1.0`
5. Duplicate `v1` code in the `solutiona/libfoo/v2` folder, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutiona/libfoo/v2.0.0`
6. Realize `Method6()` in v2 interface, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutiona/libfoo/v2.1.0`
7. Pretend to fix a bug in `Method4()` in `v1`, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutiona/libfoo/v1.1.1`
8. Cd to `path/to/solutiona/demo`, create `main.go` and add the following code:

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

9. Initialize `path/to/solutiona/demo` as a go module: **go mod init github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/demo**

10. Build: `go build`

11. Downgrade `libfoo` `v1` to `v1.0.0`: **go get github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo@v1.0.0**

12. Build again: `go build`

Key points:

1. A version is released by creating a tag and the tag MUST follow the format **<module_path>/v<Major>.<Minor>.<Patch>.** In this case, `github.com/aaronzhuo1990/blogs` is the repo URL while `golang/semantic_import_versioning/example/solutiona/libfoo` is the module path of `libfoo` `v1`. **This is the key point to make Go able to retrieve a specific version for a module.**
2. A tag can be created by using `git tag` command or creating a github release, as long as you use correct format for the tag name. For example, `v1.1.0` is created by using `git tag` command while `v1.1.1` is creating by creating a github release. You can check the [release history](https://github.com/aaronzhuo1990/blogs/tags) of this example for more details.
3. `go module init` always grabs the latest versions for the module's dependencies, which is `v1.1.1` and `v2.1.0` in this case. You need to run `**go get github.com/aaronzhuo1990/blogs/golang/semantic_import_versioning/example/solutiona/libfoo@v1.0.0**` to downgrade `v1` from `v1.1.1` to `v1.0.0` by yourself.
4. A tag or a release is essentially a snapshot for the whole repo, which is `github.com/aaronzhuo1990/blogs` in this example. However, with Go Modules, the specific versions of `v1` and `v2` can be simultaneously installed. In this case, `v1.0.0` and `v2.1.0` are simultaneously installed. You can prove this by adding `libFooV1.Method5()` in the demo and run `go build`. It will fail as `v1.0.0` does not have `Method5()`.
5. Please note that in this example, I directly committed/pushed changes into master branch just for simplifying the demo workflow. You are supposed to create a branch, commit/push changes, create a pull request and merge the changes into master branch in real development.


### Advantage of This Solution

- This solution does not require Go Modules even though it has some limitations without Go modules. Plus, this does not require you to update Go to v1.11 or later version in order to support Go modules.
- Its code organization is clear and straightforword as each `Major` version owns its codebase. This allows you to develop `v1` and `v2` very easily.
- It works well with Go modules, which means this solution allows you to convert your packages to go modules without any problems.


### Disadvantage of This Solution

- Its file structure is somehow strange. From the example, you can see that the root directory of `v1` is `path/to/solutiona/libfoo` while the root directory of `v2` is path/to/solutiona/libfoo/v2`. This somehow means `v2` lives inside `v1`. The position of `CHANGELOG.md` file also demonstrates this awkwardness.
- A lot of code is duplicated between `v1` and `v2`, as `v2` is generated by duplicating the codebase of `v1`.



## Method A: Major Branch

There is an alternative way to realize Semantic Import Versioning. But this method only works with Go Modules. The following steps demonstrate how to do it:

1. Cd to the root directory of `solutionb`.
- Comment out `Method5()` in v1 interface, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v1.0.0`
- Uncomment `Method5()` in v1 interface, commit/push changes and create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v1.1.0`
- Create a branch (say `go-semver-solutionb-libfoo-v1`) based on master branch for `v1`. We are going to use this branch instead of master branch to add features of fix bugs for `v1`.
- Switch back to master branch, update the go.mod file to include a /v2 at the end of the module path in the module directive (module semantic_import_versioning/example/solutionb/libfoo/v2)
- Modify the signature of Method5(), commit/push changes and create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v2.0.0`
- Add Method6(), commit/push changes and create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v2.1.0`
- Switch back to branch `go-semver-solutionb-libfoo-v1`, fix a bug in Method4(), commite/push changes and then create the tag `golang/semantic_import_versioning/example/solutionb/libfoo/v1.1.1` based on this branch other than master branch.


### Advantage
- It reduces a lot of code between `v1` and `v2`.
- The file structure of `v1` and `v2` makes way more sense.

### Disadvantage
- I don't know how to manage `CHANGELOG.md` file for `v1` and `v2`. It looks like it is required for having `CHANGELOG.md` for `v1` and `v2`.
- A repo may be exploded with a bunch of branches when it is managing a lot of modules. Moreover, the master branch is not unique anymore, as old `Major` versions now are using their own branches as master branch.


## Summary
- Go Modules provides a way for you to group one or more packages into a single module. While Semantic Import Versioning is a guide for introducing Semantic Versioning into Go Packages and Modules.
- There are two ways to realize Semantic Import Versioning and each of them has its own advantage and disadvantage: The first method (Major Branch) is easier and works without Go Modules and it allows you to convert your packages to modules very easily. However, it duplicates a lot of code. The second method (Major Branch) does not duplicate any code but may explode the repo since each old `Major` version needs its own `master` branch. Additionally, it only works with Go Modules.
- It is mandantory not to put `v1` into package path or module path in Go Modules. So you may want to stop using it if you are thinking of converting yours packges to go modules one day.

## What Is Next?

This blog is majorly talking about how to realize Semantic Import Versioning other than Go Modules. You may want to read [this article](https://github.com/golang/go/wiki/Modules) if you are curious about the machanism of go modules.


## The Problems We Have

### Wried Semantic Versioning

Take [website-pro-libs](https://github.com/vendasta/website-pro-libs) as an example, it is a github repo for storing all the packages(libs) used by `website-pro` micro-services and it does use semantic versioning if you look at [its release history](https://github.com/vendasta/website-pro-libs/releases). However, it does not follow the specifications of semantic versioning: It updates the `Major` version when any of its packages adds a breaking change, updates the `Minor` version when a package adds a new feature, and updates the `Patch` version when a package fixes a bug. This somehow records the release history of `website-pro-libs` repo, not individual packages. But we do not use `website-pro-libs` as a unit, we use individual packages inside the repo. Therefore, semantic versioning in this case does not help.


### No Package Versioning

Each package in the `website-pro-libs` repo uses Semantic Versioning to record its release history in a `CHANGELOG` file. However, there are two problems in current settings.

Take [siteinfo package](https://github.com/vendasta/website-pro-libs/tree/master/siteinfo) in the `website-pro-libs` repo as an example, it does use Semantic Versioning to record its release history in a [CHANGELOG.MD](https://github.com/vendasta/website-pro-libs/blob/master/siteinfo/CHANGELOG.md) file. However, there are two problems with current settings.

The first problem is there is no way to maintain old `Major` versions. You can see the `siteinfo` package has two major versions `v1` and `v2`. However, `v1` is not maintainable anymore as it lost the codebase after `v2` was released, which means you it is impossible to add new features or fix bugs for `v1`. This normally is not a huge deal if the package is only used within the organization, but it becomes a headache once the package (like a Golang SDK we wrote for clients) is consumed by the clients outsides the organization.

The second problem is there is no way to retrieve specific versions



# Package Management in Go

## What Is A Go Package


## Problems

### No Versioning

### Giant Packages



## Proposed Solution
- Break giant packages into smaller packages. This is similar to the idea behind micro-service: each package is single responsibility.
- Versioning each package




Reference:

https://akomljen.com/kubernetes-api-resources-which-group-and-version-to-use/
https://github.com/golang-standards/project-layout
[Semantic Versioning](https://semver.org/)
[Go += Package Versioning](https://research.swtch.com/vgo-intro)
[Proposal: Versioned Go Modules](https://go.googlesource.com/proposal/+/master/design/24301-versioned-go.md)
[Defining Go Modules](https://research.swtch.com/vgo-module)
[Semantic Import Versioning](https://research.swtch.com/vgo-import)
[Go Modules](https://blog.golang.org/modules2019)




Instead of designing a system that inevitably leads to large programs not building, this proposal requires that package authors follow the import compatibility rule:

If an old package and a new package have the same import path,
the new package must be backwards-compatible with the old package.

The rule is a restatement of the suggestion from the Go FAQ, quoted earlier. The quoted FAQ text ended by saying, “If a complete break is required, create a new package with a new import path.” Developers today expect to use semantic versioning to express such a break, so we integrate semantic versioning into our proposal. Specifically, major version 2 and later can be used by including the version in the path, as in:

import "github.com/go-yaml/yaml/v2"
Creating v2.0.0, which in semantic versioning denotes a major break, therefore creates a new package with a new import path, as required by import compatibility. Because each major version has a different import path, a given Go executable might contain one of each major version. This is expected and desirable. It keeps programs building and allows parts of a very large program to update from v1 to v2 independently.

`github.com/go-yaml/yaml/v2` v2 is recognized as major version. This is also accepted and recommended in go modules.


