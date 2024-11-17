# snapshot

[![License](https://img.shields.io/github/license/FollowTheProcess/snapshot)](https://github.com/FollowTheProcess/snapshot)
[![Go Reference](https://pkg.go.dev/badge/github.com/FollowTheProcess/snapshot.svg)](https://pkg.go.dev/github.com/FollowTheProcess/snapshot)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/snapshot)](https://goreportcard.com/report/github.com/FollowTheProcess/snapshot)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/snapshot?logo=github&sort=semver)](https://github.com/FollowTheProcess/snapshot)
[![CI](https://github.com/FollowTheProcess/snapshot/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/snapshot/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/snapshot/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/snapshot)

Simple, intuitive snapshot testing with Go 📸

> [!WARNING]
> **snapshot is in early development and is not yet ready for use**

![caution](./img/caution.png)

## Project Description

Snapshot testing is where you assert the result of your code is identical to a specific reference value... which is basically *all* testing. If you've ever written:

```go
if got != want {
    t.Errorf("got %v, wanted %v", got, want)
}
```

Then congratulations, you've done snapshot testing 🎉 In this case `want` is the snapshot.

The trick is, when these values get large or complicated, it's difficult to manually create and maintain the snapshot every time. The next jump up is what's typically
called "golden files".

These are files (typically manually created) that contain the expected output, any difference in what your code produces to what's in the file is an error.

Think of snapshot testing as an automated, configurable, and simple way of managing golden files. All you need to do is call `snapshot.Test(t, value)` and everything is handled for you!

## Installation

```shell
go get github.com/FollowTheProcess/snapshot@latest
```

## Quickstart

### Credits

This package was created with [copier] and the [FollowTheProcess/go_copier] project template.

[copier]: https://copier.readthedocs.io/en/stable/
[FollowTheProcess/go_copier]: https://github.com/FollowTheProcess/go_copier
