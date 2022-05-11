package calendar

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

//Calendar ..
type Calendar struct {
	ID               int64    `json:"id,omitempty"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Timezone         string   `json:"timezone"`
	Periods          []Period `json:"periods"`
	UnionCalendarIDs []int64  `json:"unionCalendarIDs,omitempty"`
	Enabled          bool     `json:"enabled"`
}

// InPeriodContains ..
type InPeriodContains struct {
	Contains bool `json:"contains"`
}

// contains check if a calendar contains a specific time (based on inclusion and exclusion periods)
// This function only checks the calendar periods (but not the unioned calendars)
// Therefore, the calendar MUST have been resolved / flatten before calling this function
func (c Calendar) contains(t time.Time) (bool, PeriodStatus, PeriodStatus, PeriodStatus) {

	statusMonth := NoInfo
	statusDay := NoInfo
	statusTime := NoInfo

	for _, period := range c.Periods {
		month, day, time := period.contains(t)
		if month == OutOfPeriod || day == OutOfPeriod || time == OutOfPeriod {
			month = OutOfPeriod
			day = OutOfPeriod
			time = OutOfPeriod
		}
		if month == InPeriod {
			statusMonth = includedToStatus(period.Included)
		}
		if day == InPeriod {
			statusDay = includedToStatus(period.Included)
		}
		if time == InPeriod {
			statusTime = includedToStatus(period.Included)
		}
	}

	status := true
	if statusMonth == NoInfo && statusDay == NoInfo && statusTime == NoInfo {
		status = false
	}
	if statusMonth == OutOfPeriod || statusDay == OutOfPeriod || statusTime == OutOfPeriod {
		status = false
	}
	return status, statusMonth, statusDay, statusTime
}

// containsWithTz check if a calendar contains a specific time (based on inclusion and exclusion periods)
// This function only checks the calendar periods (but not the unioned calendars)
// Therefore, the calendar MUST have been resolved / flatten before calling this function
func (c Calendar) containsWithTz(t time.Time) (bool, PeriodStatus, PeriodStatus, PeriodStatus) {

	statusMonth := NoInfo
	statusDay := NoInfo
	statusTime := NoInfo

	tz := time.UTC
	if c.Timezone != "" {
		var err error
		tz, err = time.LoadLocation(c.Timezone)
		if err != nil {
			zap.L().Warn("Invalid timezone", zap.Any("timezone", c.Timezone))
		}
	}
	for _, period := range c.Periods {
		month, day, time := period.containsWithTz(t, tz)
		if month == OutOfPeriod || day == OutOfPeriod || time == OutOfPeriod {
			month = OutOfPeriod
			day = OutOfPeriod
			time = OutOfPeriod
		}
		if month == InPeriod {
			statusMonth = includedToStatus(period.Included)
		}
		if day == InPeriod {
			statusDay = includedToStatus(period.Included)
		}
		if time == InPeriod {
			statusTime = includedToStatus(period.Included)
		}
	}

	status := true
	if statusMonth == NoInfo && statusDay == NoInfo && statusTime == NoInfo {
		status = false
	}
	if statusMonth == OutOfPeriod || statusDay == OutOfPeriod || statusTime == OutOfPeriod {
		status = false
	}
	return status, statusMonth, statusDay, statusTime
}

func includedToStatus(included bool) PeriodStatus {
	if included {
		return InPeriod
	}
	return OutOfPeriod
}

func getCalendar(id int64) (Calendar, bool, error) {
	calendar, found := _globalCBase.calendars[id]
	return calendar, found, nil
}

func setCalendar(calendar Calendar) {
	_globalCBase.calendars[calendar.ID] = calendar
}

// ResolveCalendar resolve a calendar definition dynamically and recursively with its subcalendars
// Resolution order :
// - For each sub-calendar (with respect of order)
//    - Sub-calendar Periods (with respect of order)
// - Calendar Periods (with respect of order)
func (c Calendar) ResolveCalendar(joinedCalendars []int64) Calendar {
	joinedCalendars = append(joinedCalendars, c.ID)
	periods := make([]Period, 0)

	// Append unioned calendars periods
	for _, unionCalendarID := range c.UnionCalendarIDs {

		var circularReference bool
		for _, id := range joinedCalendars {
			if id == unionCalendarID {
				circularReference = true
				break
			}
		}
		if circularReference {
			zap.L().Warn("Skipping Calendar union in order to avoid a circular reference", zap.Int64s("joinedCalendars", joinedCalendars), zap.Int64("unionCalendarID", unionCalendarID))
			continue
		}

		unionCalendar, found, err := getCalendar(unionCalendarID)
		if !found {
			zap.L().Warn("The calendar to join was not found", zap.Int64("calendarID", c.ID), zap.Int64("unionedCalendarID", unionCalendarID))
			continue
		}
		if err != nil {
			zap.L().Error("Cannot get calendar", zap.Int64("calendarID", c.ID), zap.Int64("unionedCalendarID", unionCalendarID), zap.Error(err))
			continue
		}

		unionCalendarResolved := unionCalendar.ResolveCalendar(joinedCalendars)
		periods = append(periods, unionCalendarResolved.Periods...)
	}

	// Append current calendar periods
	periods = append(periods, c.Periods...)

	return Calendar{
		ID:               c.ID,
		Name:             fmt.Sprintf("%s (resolved)", c.Name),
		Description:      c.Description,
		Periods:          periods,
		Enabled:          c.Enabled,
		UnionCalendarIDs: []int64{},
	}
}
