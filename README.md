# melody

[![Build Status](https://travis-ci.org/olahol/melody.svg)](https://travis-ci.org/olahol/melody)
[![Coverage Status](https://img.shields.io/coveralls/olahol/melody.svg?style=flat)](https://coveralls.io/r/olahol/melody)
[![GoDoc](https://godoc.org/github.com/olahol/melody?status.svg)](https://godoc.org/github.com/olahol/melody)

> :notes: Minimalist websocket framework for Go.

Melody is websocket framework based on [github.com/gorilla/websocket](https://github.com/gorilla/websocket)
that abstracts away the tedious parts of handling websockets. It gets out of
your way so you can write real-time apps. Features include:

* [x] Clear and easy interface similar to `net/http` or Gin.
* [x] A simple way to broadcast to all or selected connected sessions.
* [x] Message buffers making concurrent writing safe.
* [x] Automatic handling of ping/pong and session timeouts.
* [x] Store data on sessions.

## Install

```bash
go get github.com/Tooooommy/comet
```

### [examples](https://github.com/olahol/melody/tree/master/examples)

## [Documentation](https://godoc.org/github.com/olahol/melody)

## Contributors

* Ola Holmström (@olahol)
* Shogo Iwano (@shiwano)
* Matt Caldwell (@mattcaldwell)
* Heikki Uljas (@huljas)
* Robbie Trencheny (@robbiet480)
* yangjinecho (@yangjinecho)
