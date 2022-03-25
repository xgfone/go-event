// Copyright 2022 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import (
	"fmt"
	"sort"
)

func ExampleEmitter() {
	newListener := func(listenerName string) ListenerFunc {
		return func(event string, data ...interface{}) {
			fmt.Printf("listener=%s, event=%s, data=%v\n", listenerName, event, data)
		}
	}

	ln1 := newListener("ln1")
	ln2 := newListener("ln2")
	ln3 := newListener("ln3")

	On("e1", "ln1", ln1)
	On("e1", "ln2", ln2)
	On("e2", "ln2", ln2)
	OnFunc("e2", "ln3", ln3)
	OnFunc("e3", "ln3", ln3)

	events := Events()
	sort.Strings(events)
	fmt.Printf("Events: %v\n", events)

	Emit("e1", "data1")
	Emit("e2", "data2")
	Emit("e3", "data3")

	Off("e1", "ln1")
	Off("e2", "ln2")
	Off("e3", "ln3")

	events = Events()
	sort.Strings(events)
	fmt.Printf("Events: %v\n", events)

	Emit("e1", "data4")
	Emit("e2", "data5")
	Emit("e3", "data6")

	EmitAsync("e1", "data7").Wait()
	EmitAsync("e2", "data8").Wait()
	EmitAsync("e3", "data9").Wait()

	// Remove the event "e2" and all its listeners.
	Off("e2", "")

	events = Events()
	sort.Strings(events)
	fmt.Printf("Events: %v\n", events)

	// Remove all the events and their listeners.
	Off("", "")

	events = Events()
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
