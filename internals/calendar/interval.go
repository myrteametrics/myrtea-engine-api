package calendar

import (
	"time"
)

type dateTimeInterval struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func (i dateTimeInterval) containsWithTz(t time.Time, tz *time.Location) bool {
	return t.In(tz).After(i.From) && t.In(tz).Before(i.To)
}

type monthInterval struct {
	From time.Month `json:"from"`
	To   time.Month `json:"to"`
}

func (i monthInterval) containsWithTz(t time.Time, tz *time.Location) bool {
	return t.In(tz).Month() >= i.From && t.In(tz).Month() <= i.To
}

type dayInterval struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func (i dayInterval) containsWithTz(t time.Time, tz *time.Location) bool {
	return t.In(tz).Day() >= i.From && t.In(tz).Day() <= i.To
}

type dayWeekInterval struct {
	From time.Weekday `json:"from"`
	To   time.Weekday `json:"to"`
}

func (i dayWeekInterval) containsWithTz(t time.Time, tz *time.Location) bool {
	return t.In(tz).Weekday() >= i.From && t.In(tz).Weekday() <= i.To
}

type hoursInterval struct {
	FromHour   int `json:"fromHour"`
	FromMinute int `json:"fromMinute"`
	ToHour     int `json:"toHour"`
	ToMinute   int `json:"toMinute"`
}

func (i hoursInterval) containsWithTz(t time.Time, tz *time.Location) bool {
	fromMinutes := i.FromHour*60 + i.FromMinute
	toMinutes := i.ToHour*60 + i.ToMinute
	tMinutes := t.In(tz).Hour()*60 + t.In(tz).Minute()

	return tMinutes >= fromMinutes && tMinutes <= toMinutes
}
