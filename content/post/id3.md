+++
description = "The ULID ID format"
title = "Unique IDs in Golang, part 3"
date = "2017-06-03T08:15:09+03:00"
categories = [ "programming", "golang", "uid", "uuid" ]
keywords = [ "programming", "golang", "uid", "uuid" ]
draft = false
+++

> This is a continuing series on *UID* alternatives:

> - [Part1]({{<relref "post/id1.md">}}) Introduces the topic
> - [Part2]({{<relref "post/id2.md">}}) Talks about *UUID*
> - [Part3]({{<relref "post/id3.md">}}) Talks about *ULID* (this post)

As we saw on the [second part]({{<relref "post/id2.md">}}), *UUIDs* aren't
without shortcomings. An alternative that tries to tackle some of these, while
still keeping some of the advantages of both *UUIDs* versions 1 and 4 are
[ULIDs](https://github.com/alizain/ulid).

The way it attempts to solve *UUID* issues are:

- No separate versions/variants, a single format to parse and generate
- 128 bits length, like *UUID*, but without the need to encode variant and
  version, all of that is *ID* information
- [Crockford's base32](http://www.crockford.com/wrmg/base32.html) encoding. This
  means the 128 bits can be encoded in a 26 character's string, that it's
  url-safe and easy for humans to read and communicate, as it's all numbers and
  uppercase letters, no special characters. Much more compact than *UUID's* 36
  characters as well
- 80 bits of randomness
- 48 bits of timestamp (UNIX-time in milliseconds), lasts until 10895 AD
- Timestamp goes first, which means *ULIDs* are roughly time-sortable and
  produce less fragmentation than *UUIDs* in many data structures

The string representation is thus (copied from the
[oklog/ulid](https://github.com/oklog/ulid) documentation):

```lisp
   01AN4Z07BY      79KA1307SR9X4MV3
  |----------|    |----------------|
   Timestamp           Entropy
    10 chars           16 chars
    48 bits            80 bits
     base32             base32
```

So in a way, it mixes the timestamp properties of *UUID* versions 1 and 2, with
better time resolution than version 2, and swapping the unreliable and insecure
*MAC* address of it for 80 bits of randomness *a-la* *UUID* version 4. Is this
good enough? Let's find out!

## Strength

How improbable is to generate a collision with *ULIDs*? Well for each
millisecond, we have 80 bits of randomness, which equals to a space of `$2^{80}
\approx 1.21\times10^{24}$` different *ULIDs* per millisecond. Of course, here
*millisecond* means *roughly* a millisecond, specially with *ULIDs* being
generated on different nodes with clock drift and
[NTP](ihttps://en.wikipedia.org/wiki/Network_Time_Protocol) to adjust their
clocks. A discussion of that is beyond the scope of this article, so just check
[this](http://book.mixu.net/distsys/time.html) reference if you are interested.
Here we are gonna consider it close enough to a real millisecond to ignore these
matters :).

We can calculate the probability of a collision within a millisecond using the
approach to solve the [birthday
problem](https://en.wikipedia.org/wiki/Birthday_problem):

<div>
$$\sqrt{2\times2^{80} \times \ln \frac{1}{1-p}}$$
</div>

were `$p$` is the approximate probability to find a collision. So if we want to
have `$1/1,000,000,000$` (1 in a billion) probability of a *ULID* collision, we
have to generate:

<div>
$$\sqrt{2\times2^{80} \times \ln \frac{1}{1-\frac{1}{1,000,000,000}}} \approx
4.92 \times 10^{7}$$
</div>

So we have to generate over 49 million *ULIDs* in a single millisecond to have a
1 in a billion probability of a collision. Seems good enough to me.

## Golang libraries and benchmarks

So, what do we have in Golang to tap on this *ULID* goodness? There seem to be
two main alternatives:

- [oklog/ulid](https://github.com/oklog/ulid) 803 github stars at this moment,
  maintained as part of the [oklog/oklog](https://github.com/oklog/oklog)
  project
- [imdario/go-ulid](https://github.com/imdario/go-ulid) 14 stars when writing
  this, no recent changes

As with the *UUID* alternatives, normally (or in case of doubt) we would want to
use `crypto/rand` as the entropy source of these, and this is indeed the default
of `imdario/go-ulid`. There seems to be a provision to perhaps change the random
source
[though](https://github.com/imdario/go-ulid/blob/80e2735d1c2c2b1d963bccaa2306f67e5f8d5da9/ulid.go#L22)
but apparently it ended up not being implemented, so `imdario/go-ulid` just uses
`crypto/rand` in an easy to use but not configurable way.

In comparison, `oklog/ulid` takes the total opposite approach, letting the user
specify both the time and source of entropy. Thus the user has a higher burden
on usage, but can fine-tune it for his application. For example the parent
project, `oklog/oklog`, uses a separate `math/rand` RNG generator
[per-connection handling
goroutine](https://github.com/oklog/oklog/blob/649df40982d338faf87e294a3e9bcad80f34f6a9/pkg/ingest/conn.go#L48-L51),
thus achieving very fast lock-free concurrent *ULID* generation.

Benchmarking these against `pborman/uuid` with this
[code](/files/id3/ulid_generating_test.go):

```go
  package main

  import (
          crand "crypto/rand"
          mrand "math/rand"
          "testing"

          imdarioULID "github.com/imdario/go-ulid"
          oklogULID "github.com/oklog/ulid"
          pbormanUUID "github.com/pborman/uuid"
  )

  func BenchmarkPbormanUUIDV4(b *testing.B) {
          for i := 0; i < b.N; i++ {
                  pbormanUUID.NewRandom()
          }
  }

  func BenchmarkImdarioULID(b *testing.B) {
          for i := 0; i < b.N; i++ {
                  imdarioULID.New()
          }
  }

  func BenchmarkOklogULID(b *testing.B) {
          for i := 0; i < b.N; i++ {
                  oklogULID.MustNew(oklogULID.Now(), crand.Reader)
          }
  }

  func BenchmarkPbormanUUIDV4MathRand(b *testing.B) {
          rsource := mrand.New(mrand.NewSource(1)) // Unsafe concurrent use!
          pbormanUUID.SetRand(rsource)
          for i := 0; i < b.N; i++ {
                  pbormanUUID.NewRandom()
          }
  }

  func BenchmarkOklogULIDMathRand(b *testing.B) {
          rsource := mrand.New(mrand.NewSource(1)) // Unsafe concurrent use!
          for i := 0; i < b.N; i++ {
                  oklogULID.MustNew(oklogULID.Now(), rsource)
          }
  }
```

We get:

```lisp
BenchmarkPbormanUUIDV4-8           	 3000000	       493 ns/op
BenchmarkImdarioULID-8             	 3000000	       527 ns/op
BenchmarkOklogULID-8               	 3000000	       508 ns/op
BenchmarkPbormanUUIDV4MathRand-8   	20000000	        72.3 ns/op
BenchmarkOklogULIDMathRand-8       	20000000	        84.9 ns/op
```

We get that `pborman/uuid` is a bit faster than `oklog/ulid` on similar
conditions. This was counterintuitive for me, as *ULIDs* require less entropy
(80 bits) than *UUIDs V4* (122 bits). Anyway, getting the current time and
converting it to milliseconds isn't free, and overall it's on the same ballpark.

As for to/from string format, I was surprised to find that `imdario/go-ulid`
doesn't parse the *ULID* string representation, so we can only test the
to-string functionality. Benchmarking with this
[code](/files/id3/ulid_string_test.go):

```go
  package main

  import (
          "crypto/rand"
          "testing"

          imdarioULID "github.com/imdario/go-ulid"
          oklogULID "github.com/oklog/ulid"
          pbormanUUID "github.com/pborman/uuid"
  )

  var (
          UUID        = pbormanUUID.NewRandom()
          UUIDstring  = UUID.String()
          ULIDimdario = imdarioULID.New()
          ULIDoklog   = oklogULID.MustNew(oklogULID.Now(), rand.Reader)
          ULIDstring  = imdarioULID.New().String()
  )

  func BenchmarkPbormanUUIDToString(b *testing.B) {
          for i := 0; i < b.N; i++ {
                  _ = UUID.String()
          }
  }

  func BenchmarkImdarioULIDToString(b *testing.B) {
          for i := 0; i < b.N; i++ {
                  _ = ULIDimdario.String()
          }
  }

  func BenchmarkOklogULIDToString(b *testing.B) {
          for i := 0; i < b.N; i++ {
                  _ = ULIDoklog.String()
          }
  }

  func BenchmarkPbormanUUIDFromString(b *testing.B) {
          for i := 0; i < b.N; i++ {
                  pbormanUUID.Parse(UUIDstring)
          }
  }

  func BenchmarkOklogUlidFromString(b *testing.B) {
          for i := 0; i < b.N; i++ {
                  oklogULID.MustParse(ULIDstring)
          }
  }
```

Shows:

```lisp
BenchmarkPbormanUUIDToString-8     	20000000	        91.9 ns/op
BenchmarkImdarioULIDToString-8     	20000000	        83.7 ns/op
BenchmarkOklogULIDToString-8       	20000000	        72.2 ns/op
BenchmarkPbormanUUIDFromString-8   	20000000	        62.1 ns/op
BenchmarkOklogUlidFromString-8     	50000000	        28.1 ns/op
```

We can see that having a single format to parse really pays off. It's also
noteworthy that generating an *ULID* in string representation is roughly the
same than generating an *UUID* string, as it is the sum of generation plus
converting to a string.

As to which *Golang* package to use for *ULID* handling, `oklog/ulid` is the
clear choice, being much more flexible and featureful. That's normal, being the
newer and better maintained package.

## Conclusion

*ULIDs* try to tackle *UUID* shortcomings in a variety of ways. In the end, what
they bring to the table is a much more efficient string representation,
both in terms of size (26 versus 36 characters), human readability, and
parse/generation performance. They also have the propriety of being roughly
sortable to the millisecond, which can be very useful, prevents fragmentation,
and provides extra information "for free".

Considering that generating an *ULID* in string format costs about the same as
the *UUID V4* alternative, the sortable property, and the much more compact
representation, while still being easy to use, I found myself using `oklog/ulid`
quite a bit.

Hybrid sequential + random *UID* systems can offer even better performance by
reducing the amount of entropy needed for each single *UID*, at the cost of
being more complex to setup and operate. A popular system using this schema is
Twitter's [Snowflake](https://github.com/twitter/snowflake). We'll be discussing
an [Snowflake](https://github.com/twitter/snowflake) alternative,
[noeqd](https://github.com/bmizerany/noeqd), and it's *Golang* library, in part
4.

## References

- [ULID](https://github.com/alizain/ulid) Original *ULID* implementation and
  specification
- [Crockford's base32](http://www.crockford.com/wrmg/base32.html) Article
- [Birthday problem](https://en.wikipedia.org/wiki/Birthday_problem) Wikipedia
  article
- [Network Time Protocol](https://en.wikipedia.org/wiki/Network_Time_Protocol)
  Wikipedia article
- [Distributed Systems for Fun and
  Profit](http://book.mixu.net/distsys/time.html) Good time and order discussion
- [Managing Syscall Overhead with
  crypto/rand](http://blog.sgmansfield.com/2016/06/managing-syscall-overhead-with-crypto-rand/)
  is an excellent blog post on how to get better performance from `crypto/rand`
- [The hidden dangers of default
  rand](http://blog.sgmansfield.com/2016/01/the-hidden-dangers-of-default-rand/)
  another excellent post, this time about `math/rand`

<script type="text/javascript"
  src="https://cdnjs.cloudflare.com/ajax/libs/mathjax/2.7.1/MathJax.js?config=TeX-AMS-MML_HTMLorMML">
</script>
<script type="text/x-mathjax-config">
MathJax.Hub.Config({
  tex2jax: {
    inlineMath: [['$','$'], ['\\(','\\)']],
    displayMath: [['$$','$$'], ['\[','\]']],
    processEscapes: true,
    processEnvironments: true,
    skipTags: ['script', 'noscript', 'style', 'textarea', 'pre'],
    TeX: { equationNumbers: { autoNumber: "AMS" },
         extensions: ["AMSmath.js", "AMSsymbols.js"] }
  }
});
</script>

<script type="text/x-mathjax-config">
  MathJax.Hub.Queue(function() {
    // Fix <code> tags after MathJax finishes running. This is a
    // hack to overcome a shortcoming of Markdown. Discussion at
    // https://github.com/mojombo/jekyll/issues/199
    var all = MathJax.Hub.getAllJax(), i;
    for(i = 0; i < all.length; i += 1) {
        all[i].SourceElement().parentNode.className += ' has-jax';
    }
});
</script>
