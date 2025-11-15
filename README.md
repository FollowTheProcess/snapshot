# Snapshot

<p align="center">
<img src="https://github.com/FollowTheProcess/snapshot/raw/main/docs/img/logo.webp" alt="logo">
</p>

[![License](https://img.shields.io/github/license/FollowTheProcess/snapshot)](https://github.com/FollowTheProcess/snapshot)
[![Go Reference](https://pkg.go.dev/badge/go.followtheprocess.codes/snapshot.svg)](https://pkg.go.dev/go.followtheprocess.codes/snapshot)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/snapshot)](https://goreportcard.com/report/github.com/FollowTheProcess/snapshot)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/snapshot?logo=github&sort=semver)](https://github.com/FollowTheProcess/snapshot)
[![CI](https://github.com/FollowTheProcess/snapshot/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/snapshot/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/snapshot/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/snapshot)

Simple, intuitive snapshot testing for Go 📸

> [!WARNING]
> **snapshot is in early development and is not yet ready for use**

- [Snapshot](#snapshot)
  - [Project Description](#project-description)
  - [Installation](#installation)
  - [Quickstart](#quickstart)
  - [Why use `snapshot`?](#why-use-snapshot)
    - [📝 Consistent Serialisation](#-consistent-serialisation)
    - [🔄 Automatic Updating](#-automatic-updating)
    - [🗑️ Tidying Up](#️-tidying-up)
    - [🤓 Follows Go Conventions](#-follows-go-conventions)
  - [Filters](#filters)
    - [Credits](#credits)

## Project Description

Snapshot testing is where you assert the result of your code is identical to a specific reference value... which is basically *all* testing. If you've ever written:

```go
if got != want {
    t.Errorf("got %v, wanted %v", got, want)
}
```

Then congratulations, you've done snapshot testing 🎉 In this case `want` is the snapshot.

The trick is, when these values get large or complicated (imagine a complicated JSON document), it's difficult to manually create and maintain the snapshot every time.

The next jump up is what's typically called "golden files".

These are files (typically manually created) that contain the expected output, any difference in what your code produces to what's in the file is an error.

**Enter snapshot testing 📸**

Think of snapshot testing as an automated, configurable, and simple way of managing golden files. All you need to do is call `Snap` and everything is handled for you!

## Installation

```shell
go get go.followtheprocess.codes/snapshot@latest
```

## Quickstart

```go
import (
    "testing"

    "go.followtheprocess.codes/snapshot"
)

func TestSnapshot(t *testing.T) {
    snap := snapshot.New(t)

    snap.Snap([]string{"hello", "there", "this", "is", "a", "snapshot"})

    // This will store the above snapshot in testdata/snapshots/TestSnapshot.snap
    // then all future checks will compare against this snapshot
}
```

## Why use `snapshot`?

### 📝 Consistent Serialisation

By default, a snapshot looks like this:

```yaml
source: your_test.go
description: An optional description of the snapshot
expression: value.Show() # The expression that generated the snapshot
---
<your value>
```

This format was inspired by [insta], a popular snapshot testing library in rust

> [!TIP]
> If you want a different format, you can implement your own! Just implement the `snapshot.Formatter` interface and pass it in
> with the `snapshot.WithFormatter` option and you're away!

### 🔄 Automatic Updating

Let's say you've got a bunch of snapshots saved already, and you change your implementation. *All* those snapshots will now likely need to change (after you've carefully reviewed the changes and decided they are okay!)

`snapshot` lets you do this with one line of configuration, which you can set with a test flag or environment variable, or however you like:

```go
// something_test.go
import (
  "testing"

  "go.followtheprocess.codes/snapshot"
)

var update = flag.Bool("update", false, "Update snapshots automatically")

func TestSomething(t *testing.T) {
  // Tell snapshot to update everything if the -update flag was used
  snap := snapshot.New(t, snapshot.Update(*update))

  // .... rest of the test
}
```

> [!TIP]
> If you declare top level flags in a test file, you can pass them to `go test`. So in this case, `go test -update` would store `true` in the update var. You can also use environments variables and test them with `os.Getenv` e.g. `UPDATE_SNAPSHOTS=true go test`. Whatever works for you.

> [!WARNING]
> This will update *all* snapshots in one go, so make sure you run the tests normally first and check the diffs to make sure the changes are as expected

### 🗑️ Tidying Up

One criticism of snapshot testing is that if you restructure or rename your tests, you could end up with duplicated snapshots and/or messy unused ones cluttering up your repo. This is where the `Clean` option comes in:

```go
// something_test.go
import (
  "testing"

  "go.followtheprocess.codes/snapshot"
)

var clean = flag.Bool("clean", false, "Clean up unused snapshots")

func TestSomething(t *testing.T) {
  // Tell snapshot to prune the snapshots directory of unused snapshots
  snap := snapshot.New(t, snapshot.Clean(*clean))

  // .... rest of the test
}
```

This will erase all the snapshots currently managed by snapshot, and then run the tests as normal, creating the snapshots for all the new or renamed tests for the first time. The net result is a tidy snapshots directory with only what's needed

### 🤓 Follows Go Conventions

Snapshots are stored in a `snapshots` directory in the current package under `testdata` which is the canonical place to store test fixtures and other files of this kind, the go tool completely ignores `testdata` so you know these files will never impact your binary!

See `go help test`...

```plaintext
The go tool will ignore a directory named "testdata", making it available
to hold ancillary data needed by the tests.
```

The files will be named automatically after the test:

- Single tests will be given the name of the test e.g. `func TestMyThing(t *testing.T)` will produce a snapshot file of `testdata/snapshots/TestMyThing.snap.txt`
- Sub tests (including table driven tests) will use the sub test name e.g. `testdata/snapshots/TestAdd/positive_numbers.snap.txt`

> [!TIP]
> If you want to split your snapshots with more granularity, you can name your table driven cases with a `/` in them (e.g. `"Group/subtest name"`) and the directory hierarchy will be created automatically for you, completely cross platform!

## Filters

Sometimes, your snapshots might contain data that is randomly generated like UUIDs, or constantly changing like timestamps, or that might change on different platforms like filepaths, temp directory names etc.

These things make snapshot testing annoying because your snapshot changes every time or passes locally but fails in CI on a windows GitHub Actions runner (my personal bane!). `snapshot` lets you add "filters" to your configuration to solve this.

Filters are simply regex replacements that you specify ahead of time, and if any of them appear in your snapshot, they can be replaced by whatever you want!

For example:

```go
func TestWithFilters(t *testing.T) {
  // Some common ones to give you inspiration
  snap := snapshot.New(
    t,
    snapshot.Filter("(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", "[UUID]"), // Replace uuids with the literal string "[UUID]"
    snapshot.Filter(`\\([\w\d]|\.)`, "/$1"), // Replace windows file paths with unix equivalents
    snapshot.Filter(`/var/folders/\S+?/T/\S+`, "[TEMP_DIR]"), // Replace a macos temp dir with the literal string "[TEMP_DIR]"
  )
}
```

But you can imagine more:

- Unix timestamps
- Go `time.Duration` values
- Other types of uuids, ulids etc.

If you can find a regex for it, you can filter it out!

### Credits

This package was created with [copier] and the [FollowTheProcess/go-template] project template.

[copier]: https://copier.readthedocs.io/en/stable/
[FollowTheProcess/go-template]: https://github.com/FollowTheProcess/go-template
[insta]: https://crates.io/crates/insta
