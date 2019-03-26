# SSDeepViz - A Binary Visualization of the Algorithm

[![Build Status](https://travis-ci.org/joekir/ssdeepviz.svg?branch=master)](https://travis-ci.org/joekir/ssdeepviz)

## Dependencies

- [golang](https://golang.org/dl/)

## Running the web server

```
$ GO111MODULE=on COOKIE_SESSION_KEY=0x`openssl rand -hex 8` go run web.go
Listening on :8080
```

Browse to [localhost:8080](http://localhost:8080) to see it 

## Running tests

```
$ GO111MODULE=on COOKIE_SESSION_KEY=0x`openssl rand -hex 8` go test ./...
```
