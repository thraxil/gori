package main

import (
	"time"

	"github.com/nu7hatch/gouuid"
)

type Event interface {
	GetUUID() string
	GetAggregateID() string
	GetCommand() string
	GetData() string
	GetContext() string
	GetCreated() time.Time
	Hydrate(string, string, string, string, time.Time)
	Apply(*Page) *Page
}

type EventList []Event

func (el EventList) Apply() *Page {
	p := &Page{}

	for idx, event := range el {
		if idx == 0 {
			p.Slug = event.GetAggregateID()
			p.Created = event.GetCreated()
		}
		p = event.Apply(p)
	}
	return p
}

// common base for events

type StoredEvent struct {
	UUID        string
	AggregateID string
	Data        string
	Context     string
	Created     time.Time
}

func (e *StoredEvent) Hydrate(uuid, aggregateID, data, context string, created time.Time) {
	e.UUID = uuid
	e.AggregateID = aggregateID
	e.Data = data
	e.Context = context
	e.Created = created
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

func (e StoredEvent) GetCreated() time.Time {
	return e.Created
}

func newUUID() string {
	u4, _ := uuid.NewV4()
	return u4.String()
}

// SetTitleEvent -------------------------------------------------------------

type SetTitleEvent struct {
	StoredEvent
}

func CreateSetTitleEvent(aggregateID, data, context string) *SetTitleEvent {
	p := &SetTitleEvent{}
	p.Hydrate(newUUID(), aggregateID, data, context, time.Now())
	return p
}

func (e SetTitleEvent) GetCommand() string {
	return "set title"
}

func (e SetTitleEvent) Apply(page *Page) *Page {
	page.Title = e.Data
	page.Modified = e.Created
	return page
}

// SetBodyEvent -------------------------------------------------------------

type SetBodyEvent struct {
	StoredEvent
}

func CreateSetBodyEvent(aggregateID, data, context string) *SetBodyEvent {
	p := &SetBodyEvent{}
	p.Hydrate(newUUID(), aggregateID, data, context, time.Now())
	return p
}

func (e SetBodyEvent) GetCommand() string {
	return "set body"
}

func (e SetBodyEvent) Apply(page *Page) *Page {
	page.Body = e.Data
	page.Modified = e.Created
	return page
}
