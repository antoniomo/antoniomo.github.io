+++
description = "UUIDs discussion and drawbacks"
title = "Unique IDs in Golang, part 2"
date = "2017-05-21T09:03:16+03:00"
categories = [ "programming", "golang", "uid", "uuid" ]
keywords = [ "programming", "golang", "uid", "uuid" ]
draft = true
+++

> This is a continuing series on *UID* alternatives:

> - [Part1]({{<relref "post/id1.md">}}) Introduces the topic
> - [Part2]({{<relref "post/id2.md">}}) Talks about *UUID* (this post)

*Universal Unique Identifiers (UUID)* are an standard way of generating and
representing 128-bit numbers to be used as identifiers. The standard [RFC
4122](https://tools.ietf.org/html/rfc4122) define 5 different ways of generating
*UUIDs*:

- Type 1 concatenates the unique [MAC
  address](https://en.wikipedia.org/wiki/MAC_address) of the generating node
  with a 60-bit timestamp, corresponding to a single point of time and space and
  thus deemed unique. It's however a security concern as the `MAC` address can
  be traced to the generating node
- Type 2 is similar to type 1, but the timestamp is truncated to accomodate a
  "local domain" number, representing user ids, group ids, or the like. The
  reduced timestamp resolution means that type 2 is not suitable for cases were
  the *UUIDs* are issued per node/domain at a rate exceeding 1 every 7 seconds
- Type 3 deterministically hashes a predefined `name` from a `namespace` with
  `md5` to get the *UUID*. Not suitable for security credentials, but useful to
  assign an *UUID* to some entity with an already unique name in another format
- Type 5 as type 3, but uses `SHA1` as the hash, preferable over type 3, but
  still not suitable for security credentials
- Type 4 is completely random with 122 bits of entropy

For these reasons, type 4 is the most flexible, falling in the *Random UID*
category that we outlined in [part 1]({{<relref "post/id1.md">}}), and we are
gonna center the discussion on that variant.

## Golang RNG

Before comparing particular *UUID* packages, it's interesting to point out that
there are two ways to generate random numbers in the *Golang* standard library,
[crypto/rand](https://golang.org/pkg/crypto/rand/) and
[math/rand](https://golang.org/pkg/math/rand/). Without entering into details,
the main difference is that `crypto/rand` uses a cryptographically secure source
of entropy from the operating system (`/dev/urandom`) if available, or falls
back to a cryptographically secure algorithm for generating random values. This
involves system calls and isn't fast. In contrast, `math/rand` uses a faster
algorithm, but even when properly seeded it isn't suitable for security-related
uses. Since *UUID* are based on good-quality randomness, `crypto/rand` is the
logical choice for most packages, but some use cases might prefer a faster
approach.

Also note that it is possible to use some
[strategies](http://blog.sgmansfield.com/2016/06/managing-syscall-overhead-with-crypto-rand/)
to get the most of `crypto/rand`, and that naive usage of `math/rand` [isn't
optimal](http://blog.sgmansfield.com/2016/01/the-hidden-dangers-of-default-rand/),
specially in concurrent scenarios, which are the norm in most non-trivial Golang
projects.

## Golang *UUID* packages

By far, the most popular package for handling *UUID* in *Golang* is
[satori/go.uuid](https://github.com/satori/go.uuid), with 1136 stars on *Github*
at the time of writing. It supports the five *UUID* variants, it's well tested and
documented. This package uses `crypto/rand` to generate the random bits so it's
as secure as it can be, but I couldn't see how to specify my own source of
entropy so it's not easy to get more performance if needed.

A contender that is also well tested and RFC-compliant is
[pborman/uuid](https://github.com/pborman/uuid). It also uses
`crypto/rand` under the hood, but it includes a
`func SetRand(io.Reader)` function to easily set our own source of
entropy, potentially enabling us to get more performance or fine-tune it
for particular scenarios were `math/rand` is acceptable and we want more
speed.

With both packages claiming *RFC* compliance and being well tested, let's
center the comparison on performance, as for some systems generating and
parsing identifiers is part of the core functionality, so the format
being equal, this could be the deciding factor on which package to use.

### Generating Performance

For this test we'll be generating random, version 4 *UUID*. We'll use this
benchmark [code](files/uuid_generating_test.go):

```go
  package main

  import (
      "math/rand"
      "testing"

      pborman "github.com/pborman/uuid"
      satori "github.com/satori/go.uuid"
  )

  func BenchmarkSatoriV4(b *testing.B) {
      for i := 0; i < b.N; i++ {
          satori.NewV4()
      }
  }

  func BenchmarkPbormanV4(b *testing.B) {
      for i := 0; i < b.N; i++ {
          pborman.NewRandom()
      }
  }

  func BenchmarkPbormanV4MathRand(b *testing.B) {
      rsource := rand.New(rand.NewSource(1)) // Unsafe concurrent use!
      pborman.SetRand(rsource)
      for i := 0; i < b.N; i++ {
          pborman.NewRandom()
      }
  }
```

Running with `go test uuid_generating_test.go -bench=.` gives us:

```lisp
  BenchmarkSatoriV4-8              3000000           497 ns/op
  BenchmarkPbormanV4-8             3000000           499 ns/op
  BenchmarkPbormanV4MathRand-8    20000000            72.6 ns/op
```

In their default mode, they have pretty much the same performance, which makes
sense, but `pborman/uuid` can be more customizable. For the shake of comparison,
I used a `math/rand` *RNG* without any *mutex*, so really fast, naive, and
unsafe for concurrent use, but this could be ok having a separate *RNG* per
goroutine or connection, or using a
[sync/pool](https://golang.org/pkg/sync/#Pool), so it's still a useful
benchmark.

### To/From string performance

Another typical usage is to marshal/unmarshal *UUIDs* to/from string, for
example to save them into text files, logs, or databases.

Let's benchmark the two contenders for this common use case with this
[code](files/uuid_string_test.go):

```go
  package main

  import (
      "testing"

      pborman "github.com/pborman/uuid"
      satori "github.com/satori/go.uuid"
  )

  var (
      pbormanUUID = pborman.NewRandom()
      satoriUUID  = satori.NewV4()
      UUIDstring  = satoriUUID.String()
  )

  func BenchmarkSatoriToString(b *testing.B) {
      for i := 0; i < b.N; i++ {
          _ = satoriUUID.String()
      }
  }

  func BenchmarkPbormanToString(b *testing.B) {
      for i := 0; i < b.N; i++ {
          _ = pbormanUUID.String()
      }
  }

  func BenchmarkSatoriFromString(b *testing.B) {
      for i := 0; i < b.N; i++ {
          satori.FromString(UUIDstring)
      }
  }

  func BenchmarkPbormanFromString(b *testing.B) {
      for i := 0; i < b.N; i++ {
          pborman.Parse(UUIDstring)
      }
  }
```

Again running with `go test uuid_string_test.go -bench=.` gives us:

```lisp
  BenchmarkSatoriToString-8       20000000            95.2 ns/op
  BenchmarkPbormanToString-8      20000000            92.0 ns/op
  BenchmarkSatoriFromString-8     10000000           142 ns/op
  BenchmarkPbormanFromString-8    20000000            62.1 ns/op
```

Seems like parsing an *UUID* to a string is roughly equivalent, but
`satori/go.uuid` is *2x* slower than `pborman/uuid` when parsing from a string.

The culprit seems to be the function
[UnmarshalText](https://github.com/satori/go.uuid/blob/0aa62d5ddceb50dbcb909d790b5345affd3669b6/uuid.go#L230-L283)
which, among other things, supports an extra *UUID* format, `{UUID}` (sometimes
used by *Microsoft GUIDs*) that `pborman/uuid`
[Parse](https://github.com/pborman/uuid/blob/5b6ed1dd754eea46be2ec69c6e3f34fd4da85180/uuid.go#L57-L86)
doesn't, thus doing less work. But, what if we need that format? Using
`pborman/uuid` would require code like this:

```go
  func withBraceParse(s string) pborman.UUID {
      if s[0] == '{' {
          s = s[1 : len(s)-1]
      }
      return pborman.Parse(s)
  }
```

Which, when benchmarked with this [code](files/uuid_bracestring_test.go)
is still much faster than `satori/go.uuid` for this operation:

```lisp
  BenchmarkSatoriFromBraceString-8        10000000           135 ns/op
  BenchmarkPbormanFromBraceString-8       20000000            70.2 ns/op
```

So the need to support that format isn't a selling point for
`satori/go.uuid`.

In conclusion, for easy of use, start with `satori/go.uuid`. However if
you want more performance or customization, `pborman/uuid` seems to be
the way to go.

## Shortcomings of UUID.

While *UUID* usage is widespread, it isn't without shortcomings:

- Sections are hard to parse for a human. If we can't directly interpret each
  section, the dashes on the string representation add no value
- If we only care about 1 variant, we wouldn't need to encode the type
- Version 4, being fully random, produces fragmentation in many data structures,
  and aren't sortable in a meaningful way
- Version 4 are slow to generate, requiring 122 bits of good quality entropy
- It's just not the most efficient way to encode 128 bits into a string. For
  example, [base64](https://en.wikipedia.org/wiki/Base64) without padding would
  require 22 characters, not 36

## UUID Alternatives

Tackling *UUID* shortcomings, while still being simple and random-generated
*UIDs*, the most popular alternative is perhaps
[ULID](https://github.com/alizain/ulid), and a discussion of the format with
comparison of *Golang* libraries implementing it will be part three.

Hybrid sequential + random *UID* systems can offer even better performance by
reducing the amount of entropy needed for each single *UID*, at the cost of
being more complex to setup and operate. A popular system using this schema is
Twitter's [Snowflake](https://github.com/twitter/snowflake). We'll be discussing
an [Snowflake](https://github.com/twitter/snowflake) alternative,
[noeqd](https://github.com/bmizerany/noeqd), and it's *Golang* library in part 4.

## References

-   [UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier) Wikipedia article
-   [RFC 4122](https://tools.ietf.org/html/rfc4122) *UUID RFC* document
-   [Managing Syscall Overhead with
    crypto/rand](http://blog.sgmansfield.com/2016/06/managing-syscall-overhead-with-crypto-rand/)
    is an excellent blog post on how to get better performance from
    `crypto/rand`
-   [The hidden dangers of default
    rand](http://blog.sgmansfield.com/2016/01/the-hidden-dangers-of-default-rand/)
    another excellent post, this time about `math/rand`

