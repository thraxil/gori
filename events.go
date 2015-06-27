package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/nu7hatch/gouuid"
)

type Event interface {
	GetUUID() string
	GetAggregateID() string
	GetCommand() string
	GetData() string
	GetContext() string
	Hydrate(string, string, string, string)
}

type EventList []Event

type EventStore interface {
	Save(string, EventList) error
	GetEventsFor(string) EventList
	Dispatch(string) Event
}

type EventFactory func() Event

type PGEventStore struct {
	db       *sql.DB
	dispatch map[string]EventFactory
}

func NewPGEventStore(dbURL string) *PGEventStore {
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Println("can't open database")
		log.Println(err)
		os.Exit(1)
	}
	dispatch := make(map[string]EventFactory)
	dispatch["set title"] = func() Event { return &SetTitleEvent{} }
	dispatch["set body"] = func() Event { return &SetBodyEvent{} }

	return &PGEventStore{db: db, dispatch: dispatch}
}

func (s PGEventStore) Dispatch(command string) Event {
	return s.dispatch[command]()
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
