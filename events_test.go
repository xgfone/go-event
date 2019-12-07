package events

import (
	"fmt"
	"sort"
	"testing"
)

func TestListeners(t *testing.T) {
	ln1 := ListenerFunc(func(...interface{}) {})
	ln2 := ListenerFunc(func(...interface{}) {})
	ln3 := ListenerFunc(func(...interface{}) {})
	lns := Listeners{nil, ln1, nil, ln3, nil, ln2, nil}
	sort.Sort(lns)

	if lns[0] != ln1 || lns[1] != ln3 || lns[2] != ln2 {
		t.Error(ln1, ln3, ln2, lns[0], lns[1], lns[2])
	}
}

func ExampleEventEmitter() {
	ln1 := ListenerFunc(func(data ...interface{}) { fmt.Println("listener1:", data) })
	ln2 := ListenerFunc(func(data ...interface{}) { fmt.Println("listener2:", data) })
	ln3 := ListenerFunc(func(data ...interface{}) { fmt.Println("listener3:", data) })

	On("e1", ln1, ln2)
	On("e2", ln2, ln3)
	Once("e3", ln3)

	Emit("e1", "emit", "event", "e1")
	Emit("e2", "emit", "event", "e2")
	Emit("e3", "emit", "event", "e3")

	Off("e1", ln1)
	Off("e2", ln2)

	Emit("e1", "emit", "event", "e1")
	Emit("e2", "emit", "event", "e2")
	Emit("e3", "emit", "event", "e3")

	EmitAsync("e1", "emitAsync", "event", "e1").Wait()
	EmitAsync("e2", "emitAsync", "event", "e2").Wait()
	EmitAsync("e3", "emitAsync", "event", "e3").Wait()

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
