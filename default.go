package events

// DefaultEmitter is the default global Emitter.
var DefaultEmitter EventEmitter = New()

// On is equal to DefaultEmitter.On(event, listeners...).
func On(event string, listeners ...Listener) {
	DefaultEmitter.On(event, listeners...)
}

// Off is equal to DefaultEmitter.Off(event, listeners...).
func Off(event string, listeners ...Listener) {
	DefaultEmitter.Off(event, listeners...)
}

// Once is equal to DefaultEmitter.Once(event, listeners...).
func Once(event string, listeners ...Listener) {
	DefaultEmitter.Once(event, listeners...)
}

// Emit is equal to DefaultEmitter.Emit(event, data...).
func Emit(event string, data ...interface{}) {
	DefaultEmitter.Emit(event, data...)
}

// EmitAsync is equal to DefaultEmitter.EmitAsync(event, data...).
func EmitAsync(event string, data ...interface{}) Result {
	return DefaultEmitter.EmitAsync(event, data...)
}
