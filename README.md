# Snapshot

<p align="center">
<img src="https://github.com/FollowTheProcess/snapshot/raw/main/img/logo.jpg" alt="logo" width=75%>
</p>

[![License](https://img.shields.io/github/license/FollowTheProcess/snapshot)](https://github.com/FollowTheProcess/snapshot)
[![Go Reference](https://pkg.go.dev/badge/go.followtheprocess.codes/snapshot.svg)](https://pkg.go.dev/go.followtheprocess.codes/snapshot)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/snapshot)](https://goreportcard.com/report/github.com/FollowTheProcess/snapshot)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/snapshot?logo=github&sort=semver)](https://github.com/FollowTheProcess/snapshot)
[![CI](https://github.com/FollowTheProcess/snapshot/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/snapshot/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/snapshot/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/snapshot)

Simple, intuitive snapshot testing with Go 📸

> [!WARNING]
> **snapshot is in early development and is not yet stable**

- [Snapshot](#snapshot)
  - [Project Description](#project-description)
  - [Installation](#installation)
  - [Quickstart](#quickstart)
  - [Why use `snapshot`?](#why-use-snapshot)
    - [📝 Total Control over Serialisation](#-total-control-over-serialisation)
    - [🔄 Automatic Updating](#-automatic-updating)
    - [🗑️ Tidying Up](#️-tidying-up)
    - [🤓 Follows Go Conventions](#-follows-go-conventions)
  - [Serialisation Rules](#serialisation-rules)
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

    // This will store the above slice in testdata/snapshots/TestSnapshot.snap.txt
    // then all future checks will compare against this snapshot
}
```

## Why use `snapshot`?

### 📝 Total Control over Serialisation

A few other libraries are out there for snapshot testing in Go, but they typically control the serialisation for you, using a generic object dumping library. This means you get what you get and there's not much option to change it.

Not very helpful if you want your snapshots to be as readable as possible!

With `snapshot`, you have full control over how your type is serialised to the snapshot file (if you need it). You can either:

- Let `snapshot` take a best guess at how to serialise your type
  - With or without "filters" to tweak the result, see [Filters](#filters) below for more info
- Implement one of `snapshot.Snapper`, [json.Marshaler], [encoding.TextMarshaler], or [fmt.Stringer] to override how it's serialised

See [Serialisation Rules](#serialisation-rules) 👇🏻 for more info on how `snapshot` decides how to snap your value

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
>
> See an example of this [here](https://github.com/FollowTheProcess/test/blob/main/test_test.go)

## Serialisation Rules

`snapshot` deals with plain text files as snapshots, this keeps them easy to read/write for both computers and humans. But crucially, easy to diff in pull request reviews!

Because of this, it needs to know how to serialise your value (which could be basically any valid construct in Go) to plain text, so we follow a few basic rules in priority order:

- **`snapshot.Snapper`:** If your type implements the `Snapper` interface, this is preferred over all other potential serialisation, this allows you to have total control over how your type is snapshotted, do whatever you like in the `Snap` method, just return a `[]byte` that you'd like to look at in the snapshot and thats it!
- **[json.Marshaler]:** If your type implements [json.Marshaler], this will be used and the snapshot will be a valid JSON file (using `MarshalIndent` for readability)
- **[encoding.TextMarshaler]:** If your type implements [encoding.TextMarshaler], this will be used to render your value to the snapshot
- **[fmt.Stringer]:** If your type implements the [fmt.Stringer] interface, this is then used instead
- **Primitive Types:** Any primitive type in Go (`bool`, `int`, `string` etc.) is serialised according to the `%v` verb in the [fmt] package
- **Fallback:** If your type hasn't been caught by any of the above rules, we will snap it using the `GoString` mechanism (the `%#v` print verb) so e.g. a `[]string` would serialise as `[]string{"one", "two", "three"}`

> [!TIP]
> `snapshot` effectively goes through this list top to bottom to discover how to serialise your type, so mechanisms at the top are preferentially chosen over mechanisms lower down. If your snapshot doesn't look quite right, consider implementing a method higher up the list to get the behaviour you need

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

This package was created with [copier] and the [FollowTheProcess/go_copier] project template.

[copier]: https://copier.readthedocs.io/en/stable/
[FollowTheProcess/go_copier]: https://github.com/FollowTheProcess/go_copier
[fmt]: https://pkg.go.dev/fmt
[json.Marshaler]: https://pkg.go.dev/encoding/json#Marshaler
[encoding.TextMarshaler]: https://pkg.go.dev/encoding#TextMarshaler
[fmt.Stringer]: https://pkg.go.dev/fmt#Stringer
