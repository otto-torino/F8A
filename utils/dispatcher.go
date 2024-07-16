package utils

type DispatcherType struct {
	listeners map[string][]func(args ...interface{})
}

var Dispatcher *DispatcherType

func init() {
	Dispatcher = &DispatcherType{}
	Dispatcher.listeners = make(map[string][]func(args ...interface{}))
}

func (d *DispatcherType) On(eventName string, fn func(args ...interface{})) {
	if d.listeners[eventName] == nil {
		d.listeners[eventName] = make([]func(args ...interface{}), 0)
	}
	d.listeners[eventName] = append(d.listeners[eventName], fn)
}

func (d *DispatcherType) Emit(eventName string, args ...interface{}) {
	for _, fn := range d.listeners[eventName] {
		fn(args...)
	}
}

func (d *DispatcherType) Off(eventName string, fn func(args ...interface{})) {
	for i, f := range d.listeners[eventName] {
		if &f == &fn {
			d.listeners[eventName] = append(d.listeners[eventName][:i], d.listeners[eventName][i+1:]...)
		}
	}
}

func (d *DispatcherType) Clear() {
	d.listeners = make(map[string][]func(args ...interface{}))
}
