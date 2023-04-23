[![CircleCI](https://circleci.com/gh/RoboCup-SSL/ssl-quality-inspector/tree/master.svg?style=svg)](https://circleci.com/gh/RoboCup-SSL/ssl-quality-inspector/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/RoboCup-SSL/ssl-quality-inspector?style=flat-square)](https://goreportcard.com/report/github.com/RoboCup-SSL/ssl-quality-inspector)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/RoboCup-SSL/ssl-quality-inspector/pkg/vision)
[![Release](https://img.shields.io/github/release/RoboCup-SSL/ssl-quality-inspector.svg?style=flat-square)](https://github.com/RoboCup-SSL/ssl-quality-inspector/releases/latest)

# ssl-quality-inspector
Command line utility to inspect several metrics like network latency and ssl-vision detection

### Requirements

You need to install following dependencies first:

* Go

See [.circleci/config.yml](.circleci/config.yml) for compatible versions.

### Build
Build and install all binaries:

```shell
make install
```

### Run
Build and run main binary:

```shell
make run
```

### Update generated protobuf code
Generate the code for the `.proto` files after you've changed anything in a `.proto` file with:

```shell
make proto
```
