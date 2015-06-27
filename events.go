package main

import "github.com/nu7hatch/gouuid"

type Event interface {
	GetUUID() string
	GetAggregateID() string
	GetCommand() string
	GetData() string
	GetContext() string
	Hydrate(string, string, string, string)
}

type EventList []Event

type StoredEvent struct {
	UUID        string
	AggregateID string
	Data        string
	Context     string
}

func (e *StoredEvent) Hydrate(uuid, aggregateID, data, context string) {
	e.UUID = uuid
	e.AggregateID = aggregateID
	e.Data = data
	e.Context = context
}

func (e StoredEvent) GetUUID() string {
	return e.UUID
}

func (e StoredEvent) GetAggregateID() string {
	return e.AggregateID
}

func (e StoredEvent) GetData() string {
	return e.Data
}

func (e StoredEvent) GetContext() string {
	return e.Context
}

func newUUID() string {
	u4, _ := uuid.NewV4()
	return u4.String()
}

type SetTitleEvent struct {
	StoredEvent
}

func CreateSetTitleEvent(aggregateID, data, context string) *SetTitleEvent {
	p := &SetTitleEvent{}
	p.Hydrate(newUUID(), aggregateID, data, context)
	return p
}

func (e SetTitleEvent) GetCommand() string {
	return "set title"
}

type SetBodyEvent struct {
	StoredEvent
}

func CreateSetBodyEvent(aggregateID, data, context string) *SetBodyEvent {
	p := &SetBodyEvent{}
	p.Hydrate(newUUID(), aggregateID, data, context)
	return p
}

func (e SetBodyEvent) GetCommand() string {
	return "set body"
}
