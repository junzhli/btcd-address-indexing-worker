# Bitcoin address indexing worker
[![Build Status](https://travis-ci.org/junzhli/btcd-address-indexing-worker.svg?branch=dev)](https://travis-ci.org/junzhli/btcd-address-indexing-worker) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![GoDoc](https://godoc.org/github.com/junzhli/btcd-address-indexing-worker?status.svg)](https://godoc.org/github.com/junzhli/btcd-address-indexing-worker)  
Bitcoin Address Indexing Worker provides the web service [btcd-address-indexing-service](https://github.com/junzhli/btcd-address-indexing-service) with additional indexing, caching mechanisms and address relevant information lookups

## Table of Contents
- [Bitcoin address indexing worker](#bitcoin-address-indexing-worker)
  - [Table of Contents](#table-of-contents)
  - [More information](#more-information)
  - [Building and test](#building-and-test)
  - [Configuration and Run](#configuration-and-run)
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

Configuration and Run
-----
* Configurable options

**Please set environment variables with file called `.env` and place this under the current path where you run this program on shell. You can begin with file `.env.template` for reference**

| Name                  | Required | Default value   | Description                          |
|-----------------------|----------|-----------------|--------------------------------------|
| MONGO_HOST            | N        | 127.0.0.1:27017 | MongoDB Host[:Port]                  |
| MONGO_USER            | N        |                 | MongoDB User                         |
| MONGO_PASSWORD        | N        |                 | MongoDB Password                     |
| REDIS_HOST            | N        | 127.0.0.1:6379  | Redis Host[:Port]                    |
| REDIS_PASSWORD        | N        |                 | Redis protected password             |
| RABBITMQ_HOST         | N        | 127.0.0.1:5672  | RabbitMQ Host[:Port]                 |
| RABBITMQ_USER         | N        | guest           | RabbitMQ User                        |
| RABBITMQ_PASSWORD     | N        | guest           | RabbitMQ Password                    |
| BTCD_JSONRPC_HOST     | N        | 127.0.0.1:8334  | Btcd JSON-RPC Host[:Port]            |
| BTCD_JSONRPC_USER     | N        |                 |  Btcd JSON-RPC User                  |
| BTCD_JSONRPC_PASSWORD | N        |                 | Btcd JSON-RPC Password               |
| BTCD_JSONRPC_TIMEOUT  | N        | 600             | Btcd JSON-RPC Read Timeout (seconds) |

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
