package calendar

import (
	"math"
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
	From time.Month `json:"from" swaggertype:"integer"`
	To   time.Month `json:"to" swaggertype:"integer"`
}

func (i monthInterval) containsWithTz(t time.Time, tz *time.Location) bool {
	fromTr, toTr, valueTr := transpose(int(i.From), int(i.To), int(t.In(tz).Month()), 12)
	return valueTr >= fromTr && valueTr <= toTr
}

type dayInterval struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func (i dayInterval) containsWithTz(t time.Time, tz *time.Location) bool {
	fromTr, toTr, valueTr := transpose(int(i.From), int(i.To), int(t.In(tz).Day()), 31)
	return valueTr >= fromTr && valueTr <= toTr
}

type dayWeekInterval struct {
	From time.Weekday `json:"from" swaggertype:"integer"`
	To   time.Weekday `json:"to" swaggertype:"integer"`
}

func (i dayWeekInterval) containsWithTz(t time.Time, tz *time.Location) bool {
	fromTr, toTr, valueTr := transpose(int(i.From), int(i.To), int(t.In(tz).Weekday()), 7)
	return valueTr >= fromTr && valueTr <= toTr
}

type hoursInterval struct {
	FromHour   int `json:"fromHour"`
	FromMinute int `json:"fromMinute"`
	ToHour     int `json:"toHour"`
	ToMinute   int `json:"toMinute"`
}

func (i hoursInterval) containsWithTz(t time.Time, tz *time.Location) bool {
	from := i.FromHour*60 + i.FromMinute
	to := i.ToHour*60 + i.ToMinute
	value := t.In(tz).Hour()*60 + t.In(tz).Minute()

	fromTr, toTr, valueTr := transpose(from, to, value, 24.0*60.0)
	return valueTr >= fromTr && valueTr <= toTr
}

func transpose(from int, to int, value int, max int) (float64, float64, float64) {
	transposition := float64(max - from)
	return math.Mod(float64(from)+transposition, float64(max)),
		math.Mod(float64(to)+transposition, float64(max)),
		math.Mod(float64(value)+transposition, float64(max))
}
