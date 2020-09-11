package calendar

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Base ..
type Base struct {
	localMu           sync.RWMutex
	calendars         map[int64]Calendar
	resolvedCalendars map[int64]Calendar
	graph             *Graph
	lastUpdateTime    time.Time
}

// NewCalendarBase creates a new calendarsBase
func NewCalendarBase() *Base {
	return &Base{
		calendars:         map[int64]Calendar{},
		resolvedCalendars: map[int64]Calendar{},
		graph:             newGraph(),
	}
}

// InPeriodFromCalendarID ..
func (cBase *Base) InPeriodFromCalendarID(id int64, t time.Time) (bool, bool, error) {
	cBase.localMu.RLock()

	calendar, found := cBase.resolvedCalendars[id]
	if !found {
		return found, false, errors.New("calendar not found")
	}
	valid, _, _, _ := calendar.contains(t)

	cBase.localMu.RUnlock()
	return found, valid, nil
}

// GetResolved get calendar resolved from global CBase
func (cBase *Base) GetResolved(id int64) (Calendar, bool, error) {
	cBase.localMu.RLock()
	calendar, found := cBase.resolvedCalendars[id]
	cBase.localMu.RUnlock()
	return calendar, found, nil
}

// Update Updates the calendar map (read the new calendar from database)
func (cBase *Base) Update() {
	cBase.localMu.RLock()
	allCalendars, err := R().GetAll()
	if err != nil {
		zap.L().Error("Cannot update calendar base", zap.Error(err))
		return
	}

	for calendarID, calendar := range allCalendars {
		if calendar.Enabled {
			cBase.calendars[calendarID] = calendar
		} else {
			_, ok := cBase.calendars[calendarID]
			if ok {
				delete(cBase.calendars, calendarID)
			}
		}
	}

	cBase.resolvedCalendars = map[int64]Calendar{}
	cBase.graph = newGraph()

	for calendarID := range allCalendars {
		cBase.graph.addVertex(calendarID)
	}

	for _, calendar := range allCalendars {
		for _, unionCalendarID := range calendar.UnionCalendarIDs {
			cBase.graph.addEdge(calendar.ID, unionCalendarID)
		}
	}

	for calendarID, calendar := range allCalendars {
		cBase.graph.clearGraph()
		if !cBase.graph.Nodes[calendarID].isCyclic(cBase.graph) {
			cBase.resolvedCalendars[calendarID] = calendar.ResolveCalendar([]int64{})
		} else {
			zap.L().Warn("Cyclic reference detected for calendar: ", zap.Int64("calendarID", calendarID))
		}
	}

	cBase.localMu.RUnlock()
}
