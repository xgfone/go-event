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

type namedListener struct {
	Name  string
	Index uint64
	Listener
}

type namedListeners []namedListener

func (a namedListeners) Len() int           { return len(a) }
func (a namedListeners) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a namedListeners) Less(i, j int) bool { return a[i].Index < a[j].Index }

type emitter struct {
	lock sync.RWMutex
	evtm map[string]map[string]namedListener
	evtv atomic.Value
	eidx uint64
}

// New returns a new thread-safe event emitter.
func New() Emitter {
	e := &emitter{evtm: make(map[string]map[string]namedListener, 16)}
	e.storeEvents(nil)
	return e
}

func (e *emitter) Emit(event string, data ...interface{}) {
	evts := e.evtv.Load().(map[string]namedListeners)
	for _, listener := range evts[event] {
		listener.EventCallback(event, data...)
	}
}

func (e *emitter) EmitAsync(event string, data ...interface{}) Result {
	listeners := e.evtv.Load().(map[string]namedListeners)[event]

	wg := new(sync.WaitGroup)
	wg.Add(len(listeners))
	for _, listener := range listeners {
		go e.emitAsync(wg, listener, event, data...)
	}
	return wg
}

func (e *emitter) emitAsync(wg *sync.WaitGroup, listener namedListener,
	event string, data ...interface{}) {
	defer e.asyncDone(wg, event, listener.Name)
	listener.EventCallback(event, data...)
}

func (e *emitter) asyncDone(wg *sync.WaitGroup, evt string, ln string) {
	wg.Done()
	if err := recover(); err != nil {
		log.Printf("EventEmitter: listener '%s' panics on event '%s'", evt, ln)
	}
}

func (e *emitter) Events() []string {
	evts := e.evtv.Load().(map[string]namedListeners)
	events := make([]string, 0, len(evts))
	for event := range evts {
		events = append(events, event)
	}
	return events
}

func (e *emitter) Listeners(event string) map[string]Listener {
	lns := e.evtv.Load().(map[string]namedListeners)[event]
	listeners := make(map[string]Listener, len(lns))
	for _, listener := range lns {
		listeners[listener.Name] = listener.Listener
	}
	return listeners
}

func (e *emitter) storeEvents(evts map[string]namedListeners) {
	e.evtv.Store(evts)
}

func (e *emitter) updateEvents() {
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

func (e *emitter) On(event, listenerName string, listener Listener) {
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

func (e *emitter) Off(event, listenerName string) {
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

func (e *emitter) Once(event, listenerName string, listener Listener) {
	e.On(event, listenerName, newOnceListener(e, listenerName, listener))
}

func newOnceListener(emitter *emitter, lnname string, ln Listener) Listener {
	return &onceListener{emitter: emitter, lnname: lnname, listener: ln}
}

func (l *onceListener) EventCallback(event string, data ...interface{}) {
	if atomic.CompareAndSwapInt32(&l.emitted, 0, 1) {
		l.emitter.Off(event, l.lnname)
		l.listener.EventCallback(event, data...)
	}
}

type onceListener struct {
	emitter  *emitter
	emitted  int32
	lnname   string
	listener Listener
}
