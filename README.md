# replay [![Build Status](https://travis-ci.org/vinci-proxy/replay.png)](https://travis-ci.org/vinci-proxy/replay) [![GoDoc](https://godoc.org/github.com/vinci-proxy/replay?status.svg)](https://godoc.org/github.com/vinci-proxy/replay) [![API](https://img.shields.io/badge/status-stable-green.svg?style=flat)](https://godoc.org/github.com/vinci-proxy/replay) [![Coverage Status](https://coveralls.io/repos/github/vinci-proxy/replay/badge.svg?branch=master)](https://coveralls.io/github/vinci-proxy/replay?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/vinci-proxy/replay)](https://goreportcard.com/report/github.com/vinci-proxy/replay)

Replay HTTP traffic to multiple servers easily.

## Installation

```bash
go get -u gopkg.in/vinci-proxy/replay.v0
```

## API

See [godoc](https://godoc.org/github.com/vinci-proxy/replay) reference.

## Examples

#### Replay to multiple servers

```go
package main

import (
  "fmt"
  "gopkg.in/vinci-proxy/replay.v0"
  "gopkg.in/vinci-proxy/vinci.v0"
  "net/http"
)

func main() {
  vs := vinci.NewServer(vinci.ServerOptions{Host: "localhost", Port: 3100})

  handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("Replay server reached: %s => %s\n", r.RemoteAddr, r.URL.String())
    w.Write([]byte("replay server"))
  })

  srv1 := &http.Server{Addr: "localhost:3123"}
  srv2 := &http.Server{Addr: "localhost:3124"}
  srv1.Handler = handler
  srv2.Handler = handler

  go srv1.ListenAndServe()
  go srv2.ListenAndServe()

  replayer := replay.New("http://localhost:3123", "http://localhost:3124")
  replayer.SetHandler(func(err error, res *http.Response, req *http.Request) {
    if err != nil {
      fmt.Printf("Replay error: %s => %s\n", req.URL.String(), err)
      return
    }
    fmt.Printf("Replay response: %s => %d\n", req.URL.String(), res.StatusCode)
  })

  vs.Use(replayer)
  vs.Forward("http://httpbin.org")

  fmt.Printf("Server listening on port: %d\n", 3100)
  err := vs.Listen()
  if err != nil {
    fmt.Printf("Error: %s\n", err)
  }
}

```

## License

MIT
