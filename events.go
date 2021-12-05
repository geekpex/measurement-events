package main

import (
	"context"
	"fmt"
	"io"
)

type Event struct {
	From, To, Value int
}

func NewEventPublisher(output io.Writer) *EventPublisher {
	ev := &EventPublisher{
		Output: output,
	}

	for i := 0; i < 5; i++ {
		ev.uEvents[i] = make([]Event, 0, 5)
	}

	return ev
}

type EventPublisher struct {
	uEvents   [5][]Event
	times     [5][2]int
	Output    io.Writer
	NoHeaders bool
}

func (e *EventPublisher) Reset() {
	for i := 0; i < 5; i++ {
		e.uEvents[i] = e.uEvents[i][:0]
		e.times[i][0], e.times[i][1] = 0, 0
	}
}

// publishEvents writes stored events to e.output
func (e *EventPublisher) publishEvent(value int) {
	var i = value + 2

	if len(e.uEvents[i]) == 0 {
		return
	}

	// Send events
	for j := 0; j < len(e.uEvents[i]); j++ {
		// TODO: Use some other way to send string events: This causes too much allocations.
		//fmt.Fprintf(e.Output, "%d,%d,%d\n", e.uEvents[i][j].From, e.uEvents[i][j].To, e.uEvents[i][j].Value)
	}

	// Clear stored events
	if cap(e.uEvents[i]) > 5 {
		e.uEvents[i] = make([]Event, 0, 5)
	} else {
		e.uEvents[i] = e.uEvents[i][:0]
	}
}

// storeEvent stores event for future publishing
func (e *EventPublisher) storeEvent(value int) {
	var i = value + 2

	if e.times[i][0] == 0 || e.times[i][1] == 0 {
		// do not store event
		return
	}

	e.uEvents[i] = append(e.uEvents[i], Event{
		From:  e.times[i][0],
		To:    e.times[i][1],
		Value: value,
	})
	e.times[i][0], e.times[i][1] = 0, 0
}

func (e *EventPublisher) storeMeasurement(i int, m ValueTime) {
	if e.times[i][0] == 0 {
		e.times[i][0], e.times[i][1] = m.Time, m.Time
		return
	}
	e.times[i][1] = m.Time
}

// ProcessMeasurements reads ValueTime from input and publishes event if there is change
// in the measurement that matches the publish rules.
func (e *EventPublisher) ProcessMeasurements(ctx context.Context, input <-chan ValueTime) {
	var measurement ValueTime
	var lastValue, i int

	if !e.NoHeaders {
		// Print headers
		fmt.Fprint(e.Output, "Start Time,End Time,Level\n")
	}

	for {
		select {
		case <-ctx.Done():
			return
		case measurement = <-input:
			i = measurement.Value + 2

			// Handle negatives
			for ; i < 2; i++ {
				e.storeMeasurement(i, measurement)
			}
			// Handle positives
			for ; i > 2; i-- {
				e.storeMeasurement(i, measurement)
			}

			if lastValue == measurement.Value {
				continue
			}

			// Create / publish events
			switch {
			case lastValue > 0 && measurement.Value <= 0:
				for i = 1; i <= 2; i++ {
					e.storeEvent(i)
					e.publishEvent(i)
				}
			case lastValue < 0 && measurement.Value >= 0:
				for i = -1; i >= -2; i-- {
					e.storeEvent(i)
					e.publishEvent(i)
				}
			case lastValue == 2:
				e.storeEvent(2)
			case lastValue == -2:
				e.storeEvent(-2)
			}

			lastValue = measurement.Value
		}
	}
}
