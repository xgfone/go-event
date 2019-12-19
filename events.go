// Package events provides simple EventEmitter support for Go.
package events

import (
	"sort"
	"sync"
	"sync/atomic"
)

var (
	// DefaultMaxListeners is the default number of max listeners per event.
	DefaultMaxListeners = 0

	// LogWarn will print a warning when trying to add an event but the number
	// of the listeners is more than maxListeners.
	//
	// If it's nil, it does nothing.
	LogWarn func(format string, args ...interface{})
)

// Listener is used to represent the listener, which is a callback funtion
// in fact.
//
// Notice: it must be comparable.
type Listener interface {
	Callback(...interface{})
}

// Listeners is a set of the Listeners.
type Listeners []Listener

func (ls Listeners) Len() int      { return len(ls) }
func (ls Listeners) Swap(i, j int) { ls[i], ls[j] = ls[j], ls[i] }
func (ls Listeners) Less(i, j int) bool {
	if ls[j] == nil {
		return true
	}
	return false
}

// Result is used to represent the result of the asynchronous emitting.
type Result interface {
	// Wait doesn't return until all listeners have been called.
	Wait()
}

// EventEmitter  is used to manage and emit the event.
type EventEmitter interface {
	// GetMaxListeners returns the maximum number of the listeners.
	GetMaxListeners() int

	// SetMaxListeners sets the maximum number of the listeners.
	//
	// Set to zero for unlimited.
	SetMaxListeners(int)

	// EventCount returns the number of all registered events.
	EventCount() int

	// ListenerCount returns the number of all listeners registered to
	// a particular event.
	ListenerCount(event string) int

	// IsListened reports whether the event has been registered.
	IsListened(event string) bool

	// Events returns the name list of the all events.
	Events() []string

	// Listeners returns the list of the listeners registered to the event.
	Listeners(event string) Listeners

	// AddListener registers the listeners to the event.
	AddListener(event string, listeners ...Listener)

	// On is the alias of AddListener.
	On(event string, listeners ...Listener)

	// Once is the same as AddListener, but the listeners are triggered
	// only once then removed.
	Once(event string, listeners ...Listener)

	// RemoveListener removes the given listeners from the event.
	//
	// If no listeners, it will remove the whole event.
	RemoveListener(event string, listeners ...Listener)

	// Off is the alias of RemoveListener.
	Off(event string, listeners ...Listener)

	// Clear removes all events and all listeners.
	Clear()

	// Emit fires a particular event, which synchronously calls each of
	// the listeners registered for the event in the order they were registered,
	// passing the supplied arguments to each.
	Emit(event string, data ...interface{})

	// EmitAsync is the same as Emit, but triggers the listeners asynchronously.
	EmitAsync(event string, data ...interface{}) Result
}

//----------------------------------------------------------------------------
// ListenerFunc
//----------------------------------------------------------------------------

type listenerFunc struct {
	cb func(...interface{})
}

func (lf *listenerFunc) Callback(data ...interface{}) { lf.cb(data...) }

// ListenerFunc converts the function to Listener.
func ListenerFunc(f func(...interface{})) Listener { return &listenerFunc{f} }

//----------------------------------------------------------------------------
// Emitter
//----------------------------------------------------------------------------

type emitter struct {
	lock         sync.RWMutex
	evtListeners map[string]Listeners
	maxListeners int
}

// New returns a new EventEmitter
func New() EventEmitter {
	return &emitter{
		maxListeners: DefaultMaxListeners,
		evtListeners: make(map[string]Listeners, 16),
	}
}

func (e *emitter) GetMaxListeners() int {
	e.lock.RLock()
	max := e.maxListeners
	e.lock.RUnlock()
	return max
}

func (e *emitter) SetMaxListeners(n int) {
	if n < 0 {
		if LogWarn != nil {
			LogWarn("EventEmitter: MaxListeners must be positive number, tried to set '%d'", n)
		}
		return
	}

	e.lock.Lock()
	e.maxListeners = n
	e.lock.Unlock()
}

func (e *emitter) EventCount() int {
	e.lock.RLock()
	_len := len(e.evtListeners)
	e.lock.RUnlock()
	return _len
}

func (e *emitter) ListenerCount(evt string) int {
	e.lock.RLock()
	_len := len(e.evtListeners[evt])
	e.lock.RUnlock()
	return _len
}

func (e *emitter) Emit(evt string, data ...interface{}) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, listener := range e.evtListeners[evt] {
		listener.Callback(data...)
	}
}

func (e *emitter) EmitAsync(evt string, data ...interface{}) Result {
	e.lock.RLock()
	wg := new(sync.WaitGroup)
	for _, listener := range e.evtListeners[evt] {
		wg.Add(1)
		go e.emitAsync(wg, listener, data...)
	}
	e.lock.RUnlock()
	return wg
}

func (e *emitter) emitAsync(wg *sync.WaitGroup, ln Listener, data ...interface{}) {
	defer wg.Done()
	ln.Callback(data...)
}

func (e *emitter) IsListened(evt string) bool {
	e.lock.RLock()
	_, ok := e.evtListeners[evt]
	e.lock.RUnlock()
	return ok
}

func (e *emitter) Events() []string {
	names := []string{}
	e.lock.RLock()
	if _len := len(e.evtListeners); _len > 0 {
		names = make([]string, 0, _len)
		for name := range e.evtListeners {
			names = append(names, name)
		}
	}
	e.lock.RUnlock()

	return names
}

func (e *emitter) Listeners(evt string) Listeners {
	e.lock.RLock()
	listeners := append(Listeners{}, e.evtListeners[evt]...)
	e.lock.RUnlock()
	return listeners
}

func (e *emitter) Clear() {
	e.lock.Lock()
	e.evtListeners = make(map[string]Listeners, 16)
	e.lock.Unlock()
}

func (e *emitter) On(evt string, listener ...Listener) {
	e.AddListener(evt, listener...)
}

func (e *emitter) Off(evt string, listener ...Listener) {
	e.RemoveListener(evt, listener...)
}

func (e *emitter) AddListener(evt string, listener ...Listener) {
	if len(listener) == 0 {
		return
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	listeners := e.evtListeners[evt]
	if _len := len(listeners) + len(listener); e.maxListeners > 0 && _len > e.maxListeners {
		if LogWarn != nil {
			LogWarn(`EventEmitter: listeners exceeds the maximum '%d'`, _len)
		}
		return
	} else if listeners == nil {
		listeners = make(Listeners, 0, e.maxListeners)
	}

	for _, l := range listener {
		if !e.containListener(listeners, l) {
			listeners = append(listeners, l)
		}
	}
	e.evtListeners[evt] = listeners
}

func (e *emitter) containListener(listeners Listeners, listener Listener) bool {
	for _, l := range listeners {
		if l == listener {
			return true
		} else if oncelistener, ok := l.(*onceListener); ok {
			if oncelistener.listener == listener {
				return true
			}
		}
	}
	return false
}

func (e *emitter) removeListener(listeners Listeners, listener Listener) int {
	var num int
	for i, l := range listeners {
		if l == listener {
			listeners[i] = nil
			num++
		} else if oncelistener, ok := l.(*onceListener); ok {
			if oncelistener.listener == listener {
				listeners[i] = nil
				num++
			}
		}
	}
	return num
}

func (e *emitter) RemoveListener(evt string, listener ...Listener) {
	if listener == nil {
		e.RemoveAllListeners(evt)
		return
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	listeners := e.evtListeners[evt]

	var num int
	for _, l := range listener {
		num += e.removeListener(listeners, l)
	}
	if num > 0 {
		sort.Sort(listeners)
		listeners = listeners[:len(listeners)-num]
	}

	e.evtListeners[evt] = listeners
}

func (e *emitter) RemoveAllListeners(evt string) {
	e.lock.Lock()
	delete(e.evtListeners, evt)
	e.lock.Unlock()
}

type onceListener struct {
	emitted  int32
	evtname  string
	emitter  EventEmitter
	listener Listener
}

func (ol *onceListener) Callback(data ...interface{}) {
	if atomic.CompareAndSwapInt32(&ol.emitted, 0, 1) {
		ol.listener.Callback(data...)
		go ol.emitter.RemoveListener(ol.evtname, ol.listener)
	}
}

func newOnceListener(emitter EventEmitter, name string, listener Listener) Listener {
	return &onceListener{emitter: emitter, evtname: name, listener: listener}
}

func (e *emitter) Once(evt string, listener ...Listener) {
	if len(listener) == 0 {
		return
	}

	listeners := make([]Listener, len(listener))
	for i, l := range listener {
		listeners[i] = newOnceListener(e, evt, l)
	}
	e.AddListener(evt, listeners...)
}
