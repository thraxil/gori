package main

type EventFactory func() Event

type EventRegistry struct {
	dispatch map[string]EventFactory
}

func NewEventRegistry() *EventRegistry {
	return &EventRegistry{dispatch: make(map[string]EventFactory)}
}

func (r EventRegistry) Dispatch(command string) Event {
	return r.dispatch[command]()
}

func (r *EventRegistry) Register(command string, factory EventFactory) {
	r.dispatch[command] = factory
}
