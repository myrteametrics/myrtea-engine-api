package calendar

import (
	"time"
)

// Period of the calendars
type Period struct {
	Included          bool              `json:"included"`
	DateTimeIntervals *dateTimeInterval `json:"dateTimeIntervals,omitempty"`
	MonthsOfYear      *monthInterval    `json:"monthsOfYear,omitempty"`
	DaysOfMonth       *dayInterval      `json:"daysOfMonth,omitempty"`
	DaysOfWeek        *dayWeekInterval  `json:"daysOfWeek,omitempty"`
	HoursOfDay        *hoursInterval    `json:"hoursOfDay,omitempty"`
}

func (p Period) containsWithTz(t time.Time, tz *time.Location) (PeriodStatus, PeriodStatus, PeriodStatus) {

	statusMonth := NoInfo
	statusDay := NoInfo
	statusTime := NoInfo

	if p.DateTimeIntervals != nil {
		if p.DateTimeIntervals.containsWithTz(t, tz) {
			statusMonth = InPeriod
			statusDay = InPeriod
			statusTime = InPeriod
		} else {
			statusMonth = OutOfPeriod
			statusDay = OutOfPeriod
			statusTime = OutOfPeriod
		}
	}

	if p.MonthsOfYear != nil {
		if p.MonthsOfYear.containsWithTz(t, tz) {
			statusMonth = InPeriod
		} else {
			statusMonth = OutOfPeriod
		}
	}

	if p.DaysOfMonth != nil {
		if p.DaysOfMonth.containsWithTz(t, tz) {
			statusDay = InPeriod
		} else {
			statusDay = OutOfPeriod
		}
	}

	if p.DaysOfWeek != nil {
		if p.DaysOfWeek.containsWithTz(t, tz) {
			statusDay = InPeriod
		} else {
			statusDay = OutOfPeriod
		}
	}

	if p.HoursOfDay != nil {
		if p.HoursOfDay.containsWithTz(t, tz) {
			statusTime = InPeriod
		} else {
			statusTime = OutOfPeriod
		}
	}

	return statusMonth, statusDay, statusTime
}

func (p Period) contains(t time.Time) (PeriodStatus, PeriodStatus, PeriodStatus) {

	statusMonth := NoInfo
	statusDay := NoInfo
	statusTime := NoInfo

	if p.DateTimeIntervals != nil {
		if p.DateTimeIntervals.contains(t) {
			statusMonth = InPeriod
			statusDay = InPeriod
			statusTime = InPeriod
		} else {
			statusMonth = OutOfPeriod
			statusDay = OutOfPeriod
			statusTime = OutOfPeriod
		}
	}

	if p.MonthsOfYear != nil {
		if p.MonthsOfYear.contains(t) {
			statusMonth = InPeriod
		} else {
			statusMonth = OutOfPeriod
		}
	}

	if p.DaysOfMonth != nil {
		if p.DaysOfMonth.contains(t) {
			statusDay = InPeriod
		} else {
			statusDay = OutOfPeriod
		}
	}

	if p.DaysOfWeek != nil {
		if p.DaysOfWeek.contains(t) {
			statusDay = InPeriod
		} else {
			statusDay = OutOfPeriod
		}
	}

	if p.HoursOfDay != nil {
		if p.HoursOfDay.contains(t) {
			statusTime = InPeriod
		} else {
			statusTime = OutOfPeriod
		}
	}

	return statusMonth, statusDay, statusTime
}
