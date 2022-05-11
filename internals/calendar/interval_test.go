package calendar

import (
	"testing"
	"time"
)

func TestDayIntervalContainsTz(t *testing.T) {

	tzFR, _ := time.LoadLocation("Europe/Paris")
	tzUTC, _ := time.LoadLocation("UTC")

	t.Log(dayInterval{From: 2, To: 3}.containsWithTz(time.Date(2000, 1, 1, 0, 0, 0, 0, tzFR), tzFR), false)
	t.Log(dayInterval{From: 2, To: 3}.containsWithTz(time.Date(2000, 1, 2, 0, 0, 0, 0, tzFR), tzFR), true)
	t.Log(dayInterval{From: 2, To: 3}.containsWithTz(time.Date(2000, 1, 3, 0, 0, 0, 0, tzFR), tzFR), true)
	t.Log(dayInterval{From: 2, To: 3}.containsWithTz(time.Date(2000, 1, 4, 0, 0, 0, 0, tzFR), tzFR), false)
	t.Log()
	t.Log(dayInterval{From: 2, To: 3}.containsWithTz(time.Date(2000, 1, 1, 0, 0, 0, 0, tzFR), tzUTC), false)
	t.Log(dayInterval{From: 2, To: 3}.containsWithTz(time.Date(2000, 1, 2, 0, 0, 0, 0, tzFR), tzUTC), false)
	t.Log(dayInterval{From: 2, To: 3}.containsWithTz(time.Date(2000, 1, 3, 0, 0, 0, 0, tzFR), tzUTC), true)
	t.Log(dayInterval{From: 2, To: 3}.containsWithTz(time.Date(2000, 1, 4, 0, 0, 0, 0, tzFR), tzUTC), true)
	t.Log()

	t.Log(time.Date(2000, 1, 4, 0, 0, 0, 0, tzUTC))
	t.Log(time.Date(2000, 1, 4, 0, 0, 0, 0, tzUTC).In(tzFR))
	t.Log(time.Date(2000, 1, 4, 0, 0, 0, 0, tzFR).In(tzUTC))
}

func TestHoursInterval(t *testing.T) {
	tzFR, _ := time.LoadLocation("Europe/Paris")
	tzUTC, _ := time.LoadLocation("UTC")

	t.Log(hoursInterval{FromHour: 0, FromMinute: 0, ToHour: 4, ToMinute: 0}.containsWithTz(time.Date(2000, 1, 0, 0, 0, 0, 0, tzFR), tzFR), true)
	t.Log(hoursInterval{FromHour: 0, FromMinute: 0, ToHour: 4, ToMinute: 0}.containsWithTz(time.Date(2000, 1, 0, 0, 0, 0, 0, tzUTC), tzFR), true)
	t.Log(hoursInterval{FromHour: 0, FromMinute: 0, ToHour: 4, ToMinute: 0}.containsWithTz(time.Date(2000, 1, 0, 3, 30, 0, 0, tzUTC), tzFR), false)

	t.Log(hoursInterval{FromHour: 22, FromMinute: 0, ToHour: 23, ToMinute: 59}.containsWithTz(time.Date(2000, 1, 0, 22, 00, 0, 0, tzFR), tzUTC), false)
	t.Log(hoursInterval{FromHour: 22, FromMinute: 0, ToHour: 23, ToMinute: 59}.containsWithTz(time.Date(2000, 1, 0, 23, 00, 0, 0, tzFR), tzUTC), true)
	t.Log(hoursInterval{FromHour: 22, FromMinute: 0, ToHour: 23, ToMinute: 59}.containsWithTz(time.Date(2000, 2, 0, 00, 00, 0, 0, tzFR), tzUTC), true)
}
