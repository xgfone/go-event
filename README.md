# go-event [![Build Status](https://github.com/xgfone/go-event/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/go-event/actions/workflows/go.yml) [![GoDoc](https://godoc.org/github.com/xgfone/go-event?status.svg)](http://godoc.org/github.com/xgfone/go-event) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-event/master/LICENSE)

A simple event emmiter for Go `1.5+`. Inspired by [`Nodejs EventEmitter`](https://nodejs.org/api/event.html).

## Installation

```shell
$ go get -u github.com/xgfone/go-event
```

## Example

```go
package main

import (
	"fmt"
	"sort"

	"github.com/xgfone/go-event"
)

func main() {
	newListener := func(listenerName string) event.ListenerFunc {
		return func(event string, data ...interface{}) {
			fmt.Printf("listener=%s, event=%s, data=%v\n", listenerName, event, data)
		}
	}

	ln1 := newListener("ln1")
	ln2 := newListener("ln2")
	ln3 := newListener("ln3")

	event.On("e1", "ln1", ln1)
	event.On("e1", "ln2", ln2)
	event.On("e2", "ln2", ln2)
	event.OnFunc("e2", "ln3", ln3)
	event.OnFunc("e3", "ln3", ln3)

	events := event.Events()
	sort.Strings(events)
	fmt.Printf("Events: %v\n", events)

	event.Emit("e1", "data1")
	event.Emit("e2", "data2")
	event.Emit("e3", "data3")

	event.Off("e1", "ln1")
	event.Off("e2", "ln2")
	event.Off("e3", "ln3")

	events = event.Events()
	sort.Strings(events)
	fmt.Printf("Events: %v\n", events)

	event.Emit("e1", "data4")
	event.Emit("e2", "data5")
	event.Emit("e3", "data6")

	event.EmitAsync("e1", "data7").Wait()
	event.EmitAsync("e2", "data8").Wait()
	event.EmitAsync("e3", "data9").Wait()

	// Remove the event "e2" and all its listeners.
	event.Off("e2", "")

	events = event.Events()
	sort.Strings(events)
	fmt.Printf("Events: %v\n", events)

	// Remove all the events and their listeners.
	event.Off("", "")

	events = event.Events()
	sort.Strings(events)
	fmt.Printf("Events: %v\n", events)

	// Output:
	// Events: [e1 e2 e3]
	// listener=ln1, event=e1, data=[data1]
	// listener=ln2, event=e1, data=[data1]
	// listener=ln2, event=e2, data=[data2]
	// listener=ln3, event=e2, data=[data2]
	// listener=ln3, event=e3, data=[data3]
	// Events: [e1 e2]
	// listener=ln2, event=e1, data=[data4]
	// listener=ln3, event=e2, data=[data5]
	// listener=ln2, event=e1, data=[data7]
	// listener=ln3, event=e2, data=[data8]
	// Events: [e1]
	// Events: []
}
```
