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

// Package event provides a simple event emitter.
package event

// Emitter is used to manage and emit the event.
type Emitter interface {
	// Events returns the name list of the all events.
	Events() []string

	// Listeners returns all the listeners registered for the event.
	Listeners(event string) []Listener

	// On registers the listeners for the event.
	//
	// If the listener has been registered by the name, override it.
	On(event string, listener Listener)

	// Once is the same as On, but the listener are triggered only once then removed.
	Once(event string, listener Listener)

	// Off removes the listener named listenerName from the event.
	//
	// If event is empty, clear all the listeners of all the events.
	// If listenerName is empty or no longer listeners associated with the event,
	// remove the whole event.
	Off(event, listenerName string)

	// Emit fires a particular event, which synchronously calls each listener
	// registered for the event in the order they were registered.
	Emit(event string, data ...interface{})

	// EmitAsync is the same as Emit, but triggers the listeners asynchronously.
	EmitAsync(event string, data ...interface{}) Result
}

// Result is used to represent the result of the asynchronous emitting.
type Result interface {
	// Wait doesn't return until all listeners have been called.
	Wait()
}

// Listener is used to listen the event and called when the event is emitted.
type Listener interface {
	Callback(event string, data ...interface{})
	Name() string
}

// ListenerCallbackWrapper is a listener that forward the event callback
// to the wrapped callback function.
type ListenerCallbackWrapper struct {
	WrappedCallback func(ln Listener, event string, data ...interface{})
	Listener
}

// Callback implements the interface Listener, which forwards the calling
// to the wrapped callback function with the listener and the emitted event.
func (l ListenerCallbackWrapper) Callback(event string, data ...interface{}) {
	l.WrappedCallback(l.Listener, event, data...)
}

// Callback is an event function.
type Callback func(event string, data ...interface{})

// NewListener returns a new listener with the name and callback.
func NewListener(name string, callback Callback) Listener {
	if name == "" {
		panic("the listener name is empty")
	} else if callback == nil {
		panic("the listener callback is nil")
	}
	return listener{name: name, evcb: callback}
}

type listener struct {
	name string
	evcb Callback
}

func (l listener) Name() string                        { return l.name }
func (l listener) Callback(e string, d ...interface{}) { l.evcb(e, d...) }
