<img align="right" width="100" height="100" alt="DALL·E 2024-01-10 01 10 19 - the go gopher holding a jar of gochugaru" src="https://github.com/jzelinskie/gochugaru/assets/343539/67bb8a28-d425-472f-96ec-2abbe2982ed2"/>

# gochugaru

[![GoDoc](https://godoc.org/github.com/jzelinskie/gochugaru?status.svg)](https://godoc.org/github.com/jzelinskie/gochugaru)
[![Docs](https://img.shields.io/badge/docs-authzed.com-%234B4B6C "Authzed Documentation")](https://authzed.com/docs)
[![YouTube](https://img.shields.io/youtube/channel/views/UCFeSgZf0rPqQteiTQNGgTPg?color=%23F40203&logo=youtube&style=flat-square&label=YouTube "Authzed YouTube Channel")](https://www.youtube.com/channel/UCFeSgZf0rPqQteiTQNGgTPg)
[![Discord Server](https://img.shields.io/discord/844600078504951838?color=7289da&logo=discord "Discord Server")](https://authzed.com/discord)
[![Twitter](https://img.shields.io/badge/twitter-%40authzed-1D8EEE?logo=twitter "@authzed on Twitter")](https://twitter.com/authzed)


A SpiceDB client library striving to be as ergonomic as possible.

This library builds upon the official [authzed-go library], but tries to expose an interface that guides folks towards optimal performance and correctness.

[authzed-go library]: https://github.com/authzed/authzed-go

## Roadmap

- ✅ Security-obvious client constructors
- ✅ Defaults to SpiceDB's best compression method
- ✅ Check One/Many/Any/All methods
- ✅ Checks that use BulkChecks under the hood
- ✅ Flattened Relationship-type with Caveats
- ✅ Transaction-style API for Write
- ✅ Constructors for Consistency
- 🚧 Read/Delete with a RelationshipFilterBuilder
- 🔜 Request Debugging
- 🔜 Lookup Resources/Subjects
- 🔜 Read/Write Schema
- 🔜 Watch

## Examples

### Checks

```go
import gg "github.com/jzelinskie/gochugaru"

...

// Build up a set of relationships to be checked like any other slice.
var founders []Relationship
for _, founder := range []string{"jake", "joey", "jimmy"} {
  // There are various constructors for the Relationship type that can
  // trade-off allocations for legibility and understandability.
  rel, err := gg.RelationshipFromTriple("company:authzed", "founder", "user:"+founder)
  if err != nil {
    ...
  }
  founders = append(founders, rel)
}

// Various Check methods can be used to simplify common assertions.
allAreFounders, err := client.CheckAll(ctx, founders...)
if err != nil {
  ...
} else if !allAreFounders {
  ...
}
```

### Writes

```go
import gg "github.com/jzelinskie/gochugaru"

...

// You can assign gochugaru functions to variables for more terse usage.
rel := gg.MustRelationshipFromTriple

// Transactions are built up of preconditions that must or must not exist and
// the set of updates (creates, touches, or deletes) to be applied.
and relationship updates.
var txn gg.Txn
for _, rival := range []string{"joey", "jake"} {
  txn.MustNotMatch(rel("module:gochugaru", "creator", "user:"+rival).Filter())
}
txn.Touch(rel("module:gochugaru", "creator", "user:jimmy"))
txn.Touch(rel("module:gochugaru", "maintainer", "sam").WithCaveat("on_tuesday", map[string]any{"day": "wednesday"}))

writtenAt, err := client.Write(txn)
...
```
