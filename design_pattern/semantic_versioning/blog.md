# Semantic Versioning

What is included in this blog:
- An introduction of Semantic Versioning
- A discussion about how to use Semantic Versioning

## prerequisites
None

## What Is Semantic Versioning

[Semantic Versioning](https://semver.org/) is currently the most widely used version scheme in [Software Versioning](https://en.wikipedia.org/wiki/Software_versioning). It majorly uses a sequence of three digital numbers with the format `Major.Minor.Patch` with the following rules to indicate an unique status of a computer software:
- Increase the `Major` version when you make incompatible breaking changes to the software.
- Increase the `Minor` version when you add backwards-compatible features to the software.
- Increase the `Patch` version when you fix bugs in a backwards-compatible manner for the software.

The following labels (for pre-release) are also available as extensions to the `Major.Minor.Patch` format:
- `Alpha` are early versions of a software product that may not realize all of the features that are planed for the final version. They follow the format `Major.Minor.Patch-alpha.<order>` or `Major.Minor.Patch-alpha.<meta_data>`.  Alpha versions are testing versions for internal users within the organization developing the software. **It is highly recommended not to use `Alpha` versions in production as they are very unstable.**
- `Beta` are testing versions for a limited number of external users. They may not be stable and they may container some bugs. They follow the format `Major.Minor.Patch-beta.<order>` or `Major.Minor.Patch-beta.<meta_data>`. **It is also recommended not to use `Beta` versions in production.**
- `Gamma` are also testing versions but they are way more mature than `Beta` versions as they are supposed to clean up most of the bugs in `Beta` versions.  This also makes they very close to stable versions. `Gamma` versions normally are released with the format `Major.Minor.Patch-rc.<order>`, while `rc` is the abbreviation of the word `release candidate`. It has low risk to use them in production.
- `Stable` are final versions. They should strictly follow the the `Major.Minor.Patch` format.

### Expected Versioning Flow

Suppose you are using Semantic Versioning while developing a software product (say libfoo). You should have a `CHANGELOG.MD` file to record all the release history. Here is an example:
release stages:


```
# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## 0.1.0:

- 0.1.0: Initial release, add several features

```


## What Is Semantic Versioning For?


## Dependency Hell


## How to Use Semantic Versioning





http://www.ayqy.net/blog/%E8%AF%AD%E4%B9%89%E5%8C%96%E7%89%88%E6%9C%AC%E6%8E%A7%E5%88%B6%EF%BC%88semantic-versioning%EF%BC%89/
https://nodesource.com/blog/semver-a-primer/
http://www.u396.com/semver-range.html


