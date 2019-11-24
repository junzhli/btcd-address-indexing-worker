# Bitcoin address indexing worker
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)  [![GoDoc](https://godoc.org/github.com/junzhli/btcd-address-indexing-worker/utils/btcd?status.svg)](https://godoc.org/github.com/junzhli/btcd-address-indexing-worker/utils/btcd)    
Bitcoin Address Indexing Worker provides the web service `btcd-address-indexing-service` with additional indexing, caching mechanisms and address relevant information lookups

## Table of Contents
- [Bitcoin address indexing worker](#bitcoin-address-indexing-worker)
  - [Table of Contents](#table-of-contents)
  - [More information](#more-information)
  - [Building and test](#building-and-test)
  - [Run](#run)
  - [Author](#author)
  - [License](#license)

More information
-----
(Refer to [btcd-address-indexing-service](https://github.com/junzhli/btcd-address-indexing-service))

Building and test
-----

* Pulling repository

```bash
$ go get -v github.com/junzhli/btcd-address-indexing-worker
$ cd $GOPATH/src/github.com/junzhli/btcd-address-indexing-worker
```

* Build
  
```bash
$ go install
```

Run
-----

* For development

```bash
$ go run main.go
```

* For production

```bash
# Please follow build step at first
$ btcd-address-indexing-worker
```

Author
-----
Jeremy Li

License
-----
MIT License
