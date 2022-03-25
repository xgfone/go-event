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

// DefaultEmitter is the default Emitter.
var DefaultEmitter = New()

// Events is equal to DefaultEmitter.Events()
func Events() []string { return DefaultEmitter.Events() }

// Listeners is equal to DefaultEmitter.Listeners.(event).
func Listeners(event string) map[string]Listener {
	return DefaultEmitter.Listeners(event)
}

// On is equal to DefaultEmitter.On(event, listenerName, listener).
func On(event, listenerName string, listener Listener) {
	DefaultEmitter.On(event, listenerName, listener)
}

// Once is equal to DefaultEmitter.Once(event, listenerName, listener).
func Once(event, listenerName string, listener Listener) {
	DefaultEmitter.Once(event, listenerName, listener)
}

// OnceFunc is equal to DefaultEmitter.Once(event, listenerName, listener).
func OnceFunc(event, listenerName string, listener Listener) {
	DefaultEmitter.Once(event, listenerName, listener)
}

// OnFunc is equal to DefaultEmitter.On(event, listenerName, listener).
func OnFunc(event, listenerName string, listener ListenerFunc) {
	DefaultEmitter.On(event, listenerName, listener)
}

// Off is equal to DefaultEmitter.Off(event, listenerName).
func Off(event, listenerName string) {
	DefaultEmitter.Off(event, listenerName)
}

// Emit is equal to DefaultEmitter.Emit(event, data...).
func Emit(event string, data ...interface{}) {
	DefaultEmitter.Emit(event, data...)
}

// EmitAsync is equal to DefaultEmitter.EmitAsync(event, data...).
func EmitAsync(event string, data ...interface{}) Result {
	return DefaultEmitter.EmitAsync(event, data...)
}
