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
	lastUpdateTime    time.Time
}

// NewCalendarBase creates a new calendarsBase
func NewCalendarBase() *Base {
	return &Base{
		calendars:         map[int64]Calendar{},
		resolvedCalendars: map[int64]Calendar{},
	}
}

// InPeriodFromCalendarID ..
func (cBase *Base) InPeriodFromCalendarID(id int64, t time.Time) (bool, bool, error) {
	cBase.localMu.RLock()
	defer cBase.localMu.RUnlock()

	calendar, found := cBase.resolvedCalendars[id]
	if !found {
		return found, false, errors.New("calendar not found")
	}
	valid, _, _, _ := calendar.containsWithTz(t)

	return found, valid, nil
}

// GetResolved get calendar resolved from global CBase
func (cBase *Base) GetResolved(id int64) (Calendar, bool, error) {
	cBase.localMu.RLock()
	defer cBase.localMu.RUnlock()

	calendar, found := cBase.resolvedCalendars[id]
	return calendar, found, nil
}

// Update Updates the calendar map (read the new calendar from database)
func (cBase *Base) Update() {
	cBase.localMu.Lock()
	defer cBase.localMu.Unlock()

	allCalendars, err := R().GetAll()
	if err != nil {
		zap.L().Error("Cannot update calendar base", zap.Error(err))
		return
	}

	enabledCalendars := make(map[int64]Calendar)
	for calendarID, calendar := range allCalendars {
		if calendar.Enabled {
			enabledCalendars[calendarID] = calendar
		}
	}

	graph := newGraph()
	for calendarID := range allCalendars {
		graph.addVertex(calendarID)
	}

	for _, calendar := range allCalendars {
		for _, unionCalendarID := range calendar.UnionCalendarIDs {
			graph.addEdge(calendar.ID, unionCalendarID)
		}
	}

	resolvedCalendars := make(map[int64]Calendar)
	for calendarID, calendar := range allCalendars {
		graph.clearGraph()
		if !graph.Nodes[calendarID].isCyclic(graph) {
			resolvedCalendars[calendarID] = calendar.ResolveCalendar(enabledCalendars, []int64{})
		} else {
			zap.L().Warn("Cyclic reference detected for calendar: ", zap.Int64("calendarID", calendarID))
		}
	}

	cBase.calendars = enabledCalendars
	cBase.resolvedCalendars = resolvedCalendars
}
