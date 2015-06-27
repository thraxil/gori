package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type EventStore interface {
	Save(string, EventList) error
	GetEventsFor(string) EventList
	Dispatch(string) Event
}

type PGEventStore struct {
	db       *sql.DB
	registry *EventRegistry
}

func NewPGEventStore(dbURL string) *PGEventStore {
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Println("can't open database")
		log.Println(err)
		os.Exit(1)
	}
	registry := NewEventRegistry()
	registry.Register("set title", func() Event { return &SetTitleEvent{} })
	registry.Register("set body", func() Event { return &SetBodyEvent{} })

	return &PGEventStore{db: db, registry: registry}
}

func (s PGEventStore) Dispatch(command string) Event {
	return s.registry.Dispatch(command)
}

func (s *PGEventStore) Save(aggregateID string, events EventList) error {
	log.Println("PG EventStore: saving events for", aggregateID)
	if len(events) == 0 {
		// none to save
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}

	for _, event := range events {
		stmt, err := tx.Prepare(
			`insert into events (id, command, aggregate_id, event_data, event_context)
                    values($1, $2,      $3,           $4,         $5)`)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
		_, err = stmt.Exec(
			event.GetUUID(),
			event.GetCommand(),
			event.GetAggregateID(),
			event.GetData(),
			event.GetContext(),
		)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s PGEventStore) GetEventsFor(aggregateID string) EventList {
	events := make(EventList, 0)
	rows, err := s.db.Query(
		`select id, command, event_data, event_context
      from events
     where aggregate_id = $1
     order by created asc`, aggregateID)
	if err != nil {
		return events
	}
	var uuid string
	var command string
	var data string
	var context string

	for rows.Next() {
		err := rows.Scan(&uuid, &command, &data, &context)
		if err != nil {
			return events
		}
		e := s.Dispatch(command)
		e.Hydrate(uuid, aggregateID, data, context)
		events = append(events, e)
	}
	return events
}
