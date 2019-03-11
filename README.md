Gostore [![CircleCI](https://circleci.com/gh/K-Phoen/gostore.svg?style=svg&circle-token=3a5cf60e1746576891d969643fdebccde851cf7e)](https://circleci.com/gh/K-Phoen/gostore) [![Coverage Status](https://coveralls.io/repos/github/K-Phoen/gostore/badge.svg?branch=master)](https://coveralls.io/github/K-Phoen/gostore?branch=master) [![golangci](https://golangci.com/badges/github.com/K-Phoen/gostore.svg)](https://golangci.com/r/github.com/K-Phoen/gostore)
=======

Gostore is a [distributed hash table](https://en.wikipedia.org/wiki/Distributed_hash_table) implementation based on
the [SWIM protocol](https://blog.kevingomez.fr/2019/01/29/clusters-and-membership-discovering-the-swim-protocol/) and
rendezvous hashing to distribute data among nodes.

**Disclaimer:** this is a pet project, build to explore the world of DHTs and distributed systems. **Do not use it in production.**

## Features

* In-memory, with optional disk persistence using [BadgerDB](https://github.com/dgraph-io/badger)
* Time-To-Live (TTL) eviction policy
* Highly available
* Horizontally scalable

## Usage

TODO

### Running the tests

```
make tests
```

## License

This library is under the [MIT](LICENSE) license.
