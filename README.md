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
