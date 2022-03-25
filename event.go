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

// Package event provides a simple event emitter for Go.
package event

import (
	"log"
	"sort"
	"sync"
	"sync/atomic"
)

// Listener is used to listen the event and called when the event is emitted.
type Listener interface {
	EventCallback(event string, data ...interface{})
}

// ListenerFunc is a listener function.
type ListenerFunc func(event string, data ...interface{})

// EventCallback implements the interface Listener.
func (l ListenerFunc) EventCallback(event string, data ...interface{}) {
	l(event, data...)
}

// Result is used to represent the result of the asynchronous emitting.
type Result interface {
	// Wait doesn't return until all listeners have been called.
	Wait()
}

type namedListener struct {
	Name  string
	Index uint64
	Listener
}

type namedListeners []namedListener

func (a namedListeners) Len() int           { return len(a) }
func (a namedListeners) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a namedListeners) Less(i, j int) bool { return a[i].Index < a[j].Index }

// Emitter is used to manage and emit the event.
type Emitter struct {
	lock sync.RWMutex
	evtm map[string]map[string]namedListener
	evtv atomic.Value
	eidx uint64
}

// New returns a new event Emitter.
func New() *Emitter {
	e := &Emitter{evtm: make(map[string]map[string]namedListener, 16)}
	e.storeEvents(nil)
	return e
}

// Emit fires a particular event, which synchronously calls each listener
// registered for the event in the order they were registered.
func (e *Emitter) Emit(event string, data ...interface{}) {
	evts := e.evtv.Load().(map[string]namedListeners)
	for _, listener := range evts[event] {
		listener.EventCallback(event, data...)
	}
}

// EmitAsync is the same as Emit, but triggers the listeners asynchronously.
func (e *Emitter) EmitAsync(event string, data ...interface{}) Result {
	listeners := e.evtv.Load().(map[string]namedListeners)[event]

	wg := new(sync.WaitGroup)
	wg.Add(len(listeners))
	for _, listener := range listeners {
		go e.emitAsync(wg, listener, event, data...)
	}
	return wg
}

func (e *Emitter) emitAsync(wg *sync.WaitGroup, listener namedListener,
	event string, data ...interface{}) {
	defer e.asyncDone(wg, event, listener.Name)
	listener.EventCallback(event, data...)
}

func (e *Emitter) asyncDone(wg *sync.WaitGroup, evt string, ln string) {
	wg.Done()
	if err := recover(); err != nil {
		log.Printf("EventEmitter: listener '%s' panics on event '%s'", evt, ln)
	}
}

// Events returns the name list of the all events.
func (e *Emitter) Events() []string {
	evts := e.evtv.Load().(map[string]namedListeners)
	events := make([]string, 0, len(evts))
	for event := range evts {
		events = append(events, event)
	}
	return events
}

// Listeners returns all the listeners registered for the event.
func (e *Emitter) Listeners(event string) map[string]Listener {
	lns := e.evtv.Load().(map[string]namedListeners)[event]
	listeners := make(map[string]Listener, len(lns))
	for _, listener := range lns {
		listeners[listener.Name] = listener.Listener
	}
	return listeners
}

func (e *Emitter) storeEvents(evts map[string]namedListeners) {
	e.evtv.Store(evts)
}

func (e *Emitter) updateEvents() {
	evts := make(map[string]namedListeners, len(e.evtm)*2)
	for event, listeners := range e.evtm {
		lns := make(namedListeners, 0, len(listeners))
		for _, listener := range listeners {
			lns = append(lns, listener)
		}
		sort.Stable(lns)
		evts[event] = lns
	}
	e.storeEvents(evts)
}

// OnFunc is the same as On, but uses a function as the listener.
func (e *Emitter) OnFunc(event, listenerName string, listener ListenerFunc) {
	e.On(event, listenerName, listener)
}

// On registers the listeners with name for the event.
//
// If the listenerName has been registered, override it.
func (e *Emitter) On(event, listenerName string, listener Listener) {
	if event == "" {
		panic("the event is empty")
	} else if listenerName == "" {
		panic("the listener name is empty")
	} else if listener == nil {
		panic("the listener is nil")
	}

	nlistener := namedListener{
		Name:     listenerName,
		Index:    atomic.AddUint64(&e.eidx, 1),
		Listener: listener,
	}

	e.lock.Lock()
	if listeners, ok := e.evtm[event]; !ok {
		e.evtm[event] = map[string]namedListener{listenerName: nlistener}
	} else {
		listeners[listenerName] = nlistener
	}
	e.updateEvents()
	e.lock.Unlock()
}

// Off removes the listener named listenerName from the event.
//
// If event is empty, clear all the listeners of all the events.
// If listenerName is empty or no longer listeners associated with the event,
// remove the whole event.
func (e *Emitter) Off(event, listenerName string) {
	e.lock.Lock()
	if event == "" {
		for event := range e.evtm {
			delete(e.evtm, event)
		}
		e.storeEvents(nil)
	} else if listenerName == "" {
		if _, ok := e.evtm[event]; ok {
			delete(e.evtm, event)
			e.updateEvents()
		}
	} else if listeners, ok := e.evtm[event]; ok {
		if _, ok := listeners[listenerName]; ok {
			delete(listeners, listenerName)
			if len(listeners) == 0 {
				delete(e.evtm, event)
			}
			e.updateEvents()
		}
	}
	e.lock.Unlock()
}

// Once is the same as On, but the listener are triggered only once then removed.
func (e *Emitter) Once(event, listenerName string, listener Listener) {
	e.On(event, listenerName, newOnceListener(e, listenerName, listener))
}

func newOnceListener(emitter *Emitter, lnname string, ln Listener) Listener {
	return &onceListener{emitter: emitter, lnname: lnname, listener: ln}
}

func (l *onceListener) EventCallback(event string, data ...interface{}) {
	if atomic.CompareAndSwapInt32(&l.emitted, 0, 1) {
		l.emitter.Off(event, l.lnname)
		l.listener.EventCallback(event, data...)
	}
}

type onceListener struct {
	emitter  *Emitter
	emitted  int32
	lnname   string
	listener Listener
}
