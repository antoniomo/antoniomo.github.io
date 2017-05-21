+++
title = "Unique IDs in Golang, part 1"
categories = [ "programming", "golang", "uid" ]
keywords = [ "programming", "golang", "uid" ]
date = "2017-05-21T07:46:32+03:00"
description = "Introduction and sequential UIDs"
+++

Whenever we deal with entities in a software system, from user accounts to blog
posts, it's common to want to be able to refer to those with a unique,
non-ambiguous identifier, sometimes also for database storage under a unique
key.

## Sequential UIDs

On basic systems, these *Unique Identifiers (UID)* could be just a sequential or
incremental counter, where each new entity gets the next item in the sequence as
it's ID. If the system is rebooted, it will have to store the last assigned ID,
load it and keep on going.

But what happens if our system has to deal with multiple new entities
concurrently? Do we keep this counter under a mutex to serialize the sequence
generation? What if we have a distributed system and we have to create *UIDs* in
parallel, without access to a shared counter, at least not without expensive
network communications?

In these situations, sticking to purely sequential counters imposes some
limitations. For example, if we are going to operate on a single node, and/or we
count on shared memory, a *mutex* or some other form of synchronization would
work, at the cost of performance.

We could improve this by having a series of *UID* generators. For example, if our
expected concurrency is 8 workers, each one of them could have their own
generator, starting separate counters on 1 to 8. Summing 8 would get their next
sequence number without collisions.

This schema would also work in a distributed system with a fixed (and known
beforehand) number of nodes. We could even prefix it with a timestamp so that
even if the generators advance at different speeds, the generated UIDs are
somewhat sortable, at least within the
[NTP](https://en.wikipedia.org/wiki/Network_Time_Protocol) clock drift of each
node, which is typically within the 10s of *ms*, although sometimes it can go up
to half a second.

What happens if we want to be able to add/remove nodes from our system? We could
have a set number of nodes with a service that assigns the *UIDs*, and the other
nodes just request new *UIDs* to those, but that imposes expensive network
communications and is still an scalability bottleneck as adding more of these
*UID* nodes isn't trivial.

As we can see, *sequential UIDs* are perhaps not the best idea if we want to be
able to flexibly add/remove nodes to a distributed system, and they quickly stop
being *simple* when you want to deal with the above nuisances. Enter
*random-UIDs*.

## Random UIDs

In these schema, the *UIDs* are just randomly assigned from a number space much
bigger than the total number of entities we expect to handle in the entire
lifespawn of our system. The bigger the *UID* number space with respect to the
number of entities to be identified, the lower the probability of getting a
duplicate, to the point that we can totally dismiss the possibility of having
one, assuming a good quality random number generator.

Perhaps the most well known, standard way to get a *UID* following this schema
is the *Universally Unique Identifier (UUID)*, which will form the basis for
part two, where we'll be describing the format, comparing some golang libraries
to use it, and discussing why we might not want to use it all the time and some
alternatives (some hints: *ordering, locality, wasteful representation*).

Stay tuned!

## References

- [Unique Identifier](https://en.wikipedia.org/wiki/Unique_identifier) Wikipedia article
- [UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier) Wikipedia article
- [Distributed Systems for Fun and Profit](http://book.mixu.net/distsys/time.html) Good time and order discussion
