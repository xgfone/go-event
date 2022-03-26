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
	"log"
	"sort"
	"sync"
	"sync/atomic"
)

type indexListener struct {
	Index uint64
	Listener
}

type indexListeners []indexListener

func (a indexListeners) Len() int           { return len(a) }
func (a indexListeners) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a indexListeners) Less(i, j int) bool { return a[i].Index < a[j].Index }

type eventManager struct {
	matchEvent func(matchedEvent, emittedEvent string) bool
	events     map[string]indexListeners
}

func (m eventManager) Events() (events []string) {
	events = make([]string, 0, len(m.events))
	for event := range m.events {
		events = append(events, event)
	}
	return
}

func (m eventManager) Listeners(event string) (listeners []Listener) {
	ilisteners := m.events[event]
	listeners = make([]Listener, 0, len(ilisteners))
	for i, _len := 0, len(ilisteners); i < _len; i++ {
		listeners = append(listeners, ilisteners[i].Listener)
	}
	return
}

func (m eventManager) Emit(event string, data ...interface{}) {
	if m.matchEvent == nil {
		for _, listener := range m.events[event] {
			listener.Callback(event, data...)
		}
	} else {
		for matchedEvent, listeners := range m.events {
			if m.matchEvent(matchedEvent, event) {
				for _, listener := range listeners {
					listener.Callback(event, data...)
				}
			}
		}
	}
}

func (m eventManager) EmitAsync(event string, data ...interface{}) Result {
	wg := new(sync.WaitGroup)

	if m.matchEvent == nil {
		listeners := m.events[event]
		wg.Add(len(listeners))
		for _, listener := range listeners {
			go m.emitAsync(wg, listener, event, data...)
		}
	} else {
		for matchedEvent, listeners := range m.events {
			if m.matchEvent(matchedEvent, event) {
				wg.Add(len(listeners))
				for _, listener := range listeners {
					go m.emitAsync(wg, listener, event, data...)
				}
			}
		}
	}

	return wg
}

func (m eventManager) emitAsync(wg *sync.WaitGroup, listener Listener,
	event string, data ...interface{}) {
	defer m.asyncDone(wg, event, listener.Name())
	listener.Callback(event, data...)
}

func (m eventManager) asyncDone(wg *sync.WaitGroup, evt string, ln string) {
	wg.Done()
	if err := recover(); err != nil {
		log.Printf("EventEmitter: listener '%s' panics on event '%s'", evt, ln)
	}
}

type emitter struct {
	matchEvent func(matchedEvent, emittedEvent string) bool

	lock sync.RWMutex
	evtm map[string]map[string]indexListener
	evtv atomic.Value
	eidx uint64
}

// New returns a new thread-safe event emitter.
func New() Emitter { return NewCommon(nil) }

// NewCommon returns a new thread-safe common event emitter.
func NewCommon(matchEvent func(matchedEvent, emittedEvent string) bool) Emitter {
	e := &emitter{
		matchEvent: matchEvent,
		evtm:       make(map[string]map[string]indexListener, 16),
	}

	e.storeEvents(eventManager{})
	return e
}

func (e *emitter) loadEvents() eventManager   { return e.evtv.Load().(eventManager) }
func (e *emitter) storeEvents(m eventManager) { e.evtv.Store(m) }

func (e *emitter) updateEvents() {
	events := make(map[string]indexListeners, len(e.evtm)*2)
	for event, listeners := range e.evtm {
		lns := make(indexListeners, 0, len(listeners))
		for _, listener := range listeners {
			lns = append(lns, listener)
		}
		sort.Stable(lns)
		events[event] = lns
	}

	e.storeEvents(eventManager{matchEvent: e.matchEvent, events: events})
}

func (e *emitter) Events() []string {
	return e.loadEvents().Events()
}

func (e *emitter) Listeners(event string) []Listener {
	return e.loadEvents().Listeners(event)
}

func (e *emitter) Emit(event string, data ...interface{}) {
	e.loadEvents().Emit(event, data...)
}

func (e *emitter) EmitAsync(event string, data ...interface{}) Result {
	return e.loadEvents().EmitAsync(event, data...)
}

func (e *emitter) On(event string, listener Listener) {
	if event == "" {
		panic("the event is empty")
	} else if listener == nil {
		panic("the listener is nil")
	}

	lnname := listener.Name()
	ilistener := indexListener{
		Index:    atomic.AddUint64(&e.eidx, 1),
		Listener: listener,
	}

	e.lock.Lock()
	if listeners, ok := e.evtm[event]; !ok {
		e.evtm[event] = map[string]indexListener{lnname: ilistener}
	} else {
		listeners[lnname] = ilistener
	}
	e.updateEvents()
	e.lock.Unlock()
}

func (e *emitter) Off(event, listenerName string) {
	e.lock.Lock()
	if event == "" {
		for event := range e.evtm {
			delete(e.evtm, event)
		}
		e.storeEvents(eventManager{})
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

func (e *emitter) Once(event string, listener Listener) {
	e.On(event, newOnceListener(e, listener))
}

func newOnceListener(emitter *emitter, ln Listener) Listener {
	return &onceListener{emitter: emitter, Listener: ln}
}

func (l *onceListener) Callback(event string, data ...interface{}) {
	if atomic.CompareAndSwapInt32(&l.emitted, 0, 1) {
		l.emitter.Off(event, l.Name())
		l.Listener.Callback(event, data...)
	}
}

type onceListener struct {
	emitter *emitter
	emitted int32
	Listener
}
