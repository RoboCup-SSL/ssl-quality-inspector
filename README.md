[![CircleCI](https://circleci.com/gh/RoboCup-SSL/ssl-quality-inspector/tree/master.svg?style=svg)](https://circleci.com/gh/RoboCup-SSL/ssl-quality-inspector/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/RoboCup-SSL/ssl-quality-inspector?style=flat-square)](https://goreportcard.com/report/github.com/RoboCup-SSL/ssl-quality-inspector)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/RoboCup-SSL/ssl-quality-inspector/pkg/vision)
[![Release](https://img.shields.io/github/release/RoboCup-SSL/ssl-quality-inspector.svg?style=flat-square)](https://github.com/RoboCup-SSL/ssl-quality-inspector/releases/latest)

# ssl-quality-inspector
Command line utility to inspect several metrics like network latency and ssl-vision detection

## Requirements
You need to install following dependencies first: 
 * Go >= 1.11
 
## Installation

Use go get to install all packages / executables:

```
go get -u github.com/RoboCup-SSL/ssl-quality-inspector/...
```

## Run
The executables are installed to your $GOPATH/bin folder. If you have it on your $PATH, you can directly run them. 
Else, switch to this folder first.

The quality inspector can be run by:
```
ssl-quality-inspector
```

Available parameters can be retrieved with the `-h` option.
