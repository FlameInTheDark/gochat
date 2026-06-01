package goclient

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type eventHandlerInstance struct {
	id        uint64
	eventType reflect.Type
	handler   func(*Session, any) bool
	once      bool
}

var sessionType = reflect.TypeOf(&Session{})

// AddHandler adds an event handler and returns a function that removes it.
//
// Handlers must be functions with this shape:
//
//	func(*goclient.Session, *goclient.MessageCreate)
//
// The second parameter may be any concrete event pointer, *Event, or any.
func (s *Session) AddHandler(handler any) func() {
	return s.addHandler(handler, false)
}

// AddHandlerOnce adds an event handler that is removed after its first call.
func (s *Session) AddHandlerOnce(handler any) func() {
	return s.addHandler(handler, true)
}

func (s *Session) addHandler(handler any, once bool) func() {
	instance, ok := handlerForInterface(handler)
	if !ok {
		return func() {}
	}
	instance.id = atomic.AddUint64(&s.nextHandlerID, 1)
	instance.once = once
	s.handlersMu.Lock()
	s.handlers = append(s.handlers, instance)
	s.handlersMu.Unlock()
	var removeOnce sync.Once
	return func() {
		removeOnce.Do(func() {
			s.removeHandler(instance)
		})
	}
}

func handlerForInterface(handler any) (eventHandlerInstance, bool) {
	value := reflect.ValueOf(handler)
	if !value.IsValid() || value.Kind() != reflect.Func {
		return eventHandlerInstance{}, false
	}
	typ := value.Type()
	if typ.NumIn() != 2 || typ.In(0) != sessionType {
		return eventHandlerInstance{}, false
	}
	eventType := typ.In(1)
	return eventHandlerInstance{
		eventType: eventType,
		handler: func(s *Session, event any) bool {
			if event == nil {
				return false
			}
			ev := reflect.ValueOf(event)
			if !ev.Type().AssignableTo(eventType) {
				return false
			}
			value.Call([]reflect.Value{reflect.ValueOf(s), ev})
			return true
		},
	}, true
}

func (s *Session) removeHandler(target eventHandlerInstance) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	for i := range s.handlers {
		if s.handlers[i].id == target.id {
			s.handlers = append(s.handlers[:i], s.handlers[i+1:]...)
			return
		}
	}
}

func (s *Session) handle(event any) {
	s.handlersMu.RLock()
	handlers := append([]eventHandlerInstance(nil), s.handlers...)
	s.handlersMu.RUnlock()
	for _, handler := range handlers {
		called := handler.handler(s, event)
		if called && handler.once {
			s.removeHandler(handler)
		}
	}
}
