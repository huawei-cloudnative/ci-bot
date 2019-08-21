# ci-bot

**Deprecation Notice: this project is out of maintenance, please use [prow](https://github.com/kubernetes/test-infra/tree/master/prow) instead.**


[![Go Report Card](https://goreportcard.com/badge/github.com/huawei-cloudnative/ci-bot)](https://goreportcard.com/badge/github.com/huawei-cloudnative/ci-bot)
[![Build Status](https://travis-ci.org/huawei-cloudnative/ci-bot.svg?branch=master)](https://travis-ci.org/huawei-cloudnative/ci-bot)
[![LICENSE](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/huawei-cloudnative/ci-bot/blob/master/LICENSE)

This repository houses CI robot which will run your tests, add your labels, and merge your codes.

## Build

Before you get started, make sure to have [Go](https://golang.org/) already installed in your local machine.

```
$ mkdir -p $GOPATH/src/github.com/huawei-cloudnative
$ cd $GOPATH/src/github.com/huawei-cloudnative
$ git clone https://github.com/huawei-cloudnative/ci-bot
$ cd ci-bot
$ make
```

## Usage

```
$ ./ci-bot
```

## License

See the [LICENSE](LICENSE) file for details.
