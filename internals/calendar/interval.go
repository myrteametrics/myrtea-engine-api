package calendar

import (
	"time"
)

type dateTimeInterval struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func (i dateTimeInterval) contains(t time.Time) bool {
	return t.After(i.From) && t.Before(i.To)
}

type monthInterval struct {
	From time.Month `json:"from"`
	To   time.Month `json:"to"`
}

func (i monthInterval) contains(t time.Time) bool {
	return t.Month() >= i.From && t.Month() <= i.To
}

type dayInterval struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func (i dayInterval) contains(t time.Time) bool {
	return t.Day() >= i.From && t.Day() <= i.To
}

type dayWeekInterval struct {
	From time.Weekday `json:"from"`
	To   time.Weekday `json:"to"`
}

func (i dayWeekInterval) contains(t time.Time) bool {
	return t.Weekday() >= i.From && t.Weekday() <= i.To
}

type hoursInterval struct {
	FromHour   int `json:"fromHour"`
	FromMinute int `json:"fromMinute"`
	ToHour     int `json:"toHour"`
	ToMinute   int `json:"toMinute"`
}

func (i hoursInterval) contains(t time.Time) bool {
	fromMinutes := i.FromHour*60 + i.FromMinute
	toMinutes := i.ToHour*60 + i.ToMinute
	tMinutes := t.Hour()*60 + t.Minute()

	return tMinutes >= fromMinutes && tMinutes <= toMinutes
}
