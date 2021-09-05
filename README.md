# go-events [![Build Status](https://github.com/xgfone/go-events/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/go-events/actions/workflows/go.yml) [![GoDoc](https://godoc.org/github.com/xgfone/go-events?status.svg)](http://godoc.org/github.com/xgfone/go-events) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-events/master/LICENSE)

A simple event emmiter for Go `1.5+`. Inspired by [`Nodejs EventEmitter`](https://nodejs.org/api/events.html).

## Installation

```shell
$ go get -u github.com/xgfone/go-events
```

## Example

```go
package main

import (
    "fmt"

    "github.com/xgfone/go-events"
)

func main() {
    ln1 := events.ListenerFunc(func(data ...interface{}) { fmt.Println("listener1:", data) })
    ln2 := events.ListenerFunc(func(data ...interface{}) { fmt.Println("listener2:", data) })
    ln3 := events.ListenerFunc(func(data ...interface{}) { fmt.Println("listener3:", data) })

    emitter := events.New()
    emitter.On("e1", ln1, ln2)
    emitter.On("e2", ln2, ln3)
    emitter.Once("e3", ln3)

    emitter.Emit("e1", "emit", "event", "e1")
    emitter.Emit("e2", "emit", "event", "e2")
    emitter.Emit("e3", "emit", "event", "e3")

    emitter.Off("e1", ln1)
    emitter.Off("e2", ln2)

    emitter.Emit("e1", "emit", "event", "e1")
    emitter.Emit("e2", "emit", "event", "e2")
    emitter.Emit("e3", "emit", "event", "e3")

    emitter.EmitAsync("e1", "emitAsync", "event", "e1").Wait()
    emitter.EmitAsync("e2", "emitAsync", "event", "e2").Wait()
    emitter.EmitAsync("e3", "emitAsync", "event", "e3").Wait()

    // Output:
    // listener1: [emit event e1]
    // listener2: [emit event e1]
    // listener2: [emit event e2]
    // listener3: [emit event e2]
    // listener3: [emit event e3]
    // listener2: [emit event e1]
    // listener3: [emit event e2]
    // listener2: [emitAsync event e1]
    // listener3: [emitAsync event e2]
}
```

There is a default global `EventEmitter` and some method aliases, such as `On`, `Off`, `Once`, `Emit` and `EmitAsync`.

```go
package main

import (
    "fmt"

    "github.com/xgfone/go-events"
)

func main() {
    ln1 := events.ListenerFunc(func(data ...interface{}) { fmt.Println("listener1:", data) })
    ln2 := events.ListenerFunc(func(data ...interface{}) { fmt.Println("listener2:", data) })
    ln3 := events.ListenerFunc(func(data ...interface{}) { fmt.Println("listener3:", data) })

    events.On("e1", ln1, ln2)
    events.On("e2", ln2, ln3)
    events.Once("e3", ln3)

    events.Emit("e1", "emit", "event", "e1")
    events.Emit("e2", "emit", "event", "e2")
    events.Emit("e3", "emit", "event", "e3")

    events.Off("e1", ln1)
    events.Off("e2", ln2)

    events.Emit("e1", "emit", "event", "e1")
    events.Emit("e2", "emit", "event", "e2")
    events.Emit("e3", "emit", "event", "e3")

    events.EmitAsync("e1", "emitAsync", "event", "e1").Wait()
    events.EmitAsync("e2", "emitAsync", "event", "e2").Wait()
    events.EmitAsync("e3", "emitAsync", "event", "e3").Wait()

    // Output:
    // listener1: [emit event e1]
    // listener2: [emit event e1]
    // listener2: [emit event e2]
    // listener3: [emit event e2]
    // listener3: [emit event e3]
    // listener2: [emit event e1]
    // listener3: [emit event e2]
    // listener2: [emitAsync event e1]
    // listener3: [emitAsync event e2]
}
```
