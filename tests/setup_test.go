package tests

import (
	"webtyp.com/events"
	"webtyp.com/fmt"
)

type MockPublisher struct {
	Events []events.Event
}

func (m *MockPublisher) Publish(e events.Event) {
	m.Events = append(m.Events, e)
}

var _ events.Publisher = (*MockPublisher)(nil)

type MockIDGen struct {
	counter int
}

func (m *MockIDGen) NewID() string {
	m.counter++
	return "test-id-" + fmt.Convert(m.counter).String()
}
