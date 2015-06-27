package main

import (
	"log"
)

// EventStore -----------------------------------------------------

type EventStoreRepo struct {
	es EventStore
}

func NewEventStoreRepo(e EventStore) *EventStoreRepo {
	return &EventStoreRepo{es: e}
}

func (er *EventStoreRepo) FindBySlug(slug string) (*Page, error) {
	events := er.es.GetEventsFor(slug)
	log.Println("events:", len(events))
	for _, event := range events {
		log.Println("\t", event.GetCommand(), event.GetAggregateID())
	}
	page := events.Apply()
	return page, nil
}

func (er *EventStoreRepo) SetTitle(page *Page, title string) error {
	events := make(EventList, 0)
	if page.SetTitle(title) {
		events = append(events, CreateSetTitleEvent(page.Slug, page.Title, ""))
	}
	return er.es.Save(page.Slug, events)
}

func (er *EventStoreRepo) SetBody(page *Page, body string) error {
	events := make(EventList, 0)
	if page.SetBody(body) {
		events = append(events, CreateSetBodyEvent(page.Slug, page.Body, ""))
	}
	return er.es.Save(page.Slug, events)
}
