package calendar

import (
	"fmt"
	"testing"
	"time"
)

func initCalendars() {
	InitUnitTest()
	// Load calendars in cache
	setCalendar(Calendar{ID: 1, Name: "c1", Periods: []Period{{Included: true, DaysOfMonth: &dayInterval{From: 1, To: 1}}}, UnionCalendarIDs: []int64{}})
	setCalendar(Calendar{ID: 2, Name: "c2", Periods: []Period{{Included: false, DaysOfMonth: &dayInterval{From: 2, To: 2}}}, UnionCalendarIDs: []int64{}})
	setCalendar(Calendar{ID: 3, Name: "c3", Periods: []Period{{Included: true, DaysOfMonth: &dayInterval{From: 3, To: 3}}}, UnionCalendarIDs: []int64{}})
	setCalendar(Calendar{ID: 4, Name: "c4", Periods: []Period{{Included: true, DaysOfMonth: &dayInterval{From: 4, To: 4}}}, UnionCalendarIDs: []int64{2, 3}})
	setCalendar(Calendar{ID: 5, Name: "c5", Periods: []Period{{Included: true, DaysOfMonth: &dayInterval{From: 5, To: 5}}}, UnionCalendarIDs: []int64{1, 4}})
	setCalendar(Calendar{ID: 6, Name: "c5", Periods: []Period{{Included: true, DaysOfMonth: &dayInterval{From: 5, To: 5}}}, UnionCalendarIDs: []int64{1, 4, 7}})
	setCalendar(Calendar{ID: 7, Name: "c5", Periods: []Period{{Included: true, DaysOfMonth: &dayInterval{From: 5, To: 5}}}, UnionCalendarIDs: []int64{1, 4, 8}})
	setCalendar(Calendar{ID: 8, Name: "c5", Periods: []Period{{Included: true, DaysOfMonth: &dayInterval{From: 5, To: 5}}}, UnionCalendarIDs: []int64{1, 4, 6}})
}

func TestCalendarResolution(t *testing.T) {
	initCalendars()
	c, _, _ := getCalendar(5)
	resolved := c.ResolveCalendar([]int64{})

	_ = resolved
	t.Log(resolved)

}

func TestCalendarResolutionCircularReference(t *testing.T) {
	initCalendars()
	c, _, _ := getCalendar(6)
	resolved := c.ResolveCalendar([]int64{})

	_ = resolved
	t.Log(resolved)
}

func checkCalendarPeriod(t *testing.T, c Calendar, ti time.Time, mustBeInPeriod bool) error {
	status, statusMonth, statusDay, statusTime := c.containsWithTz(ti)
	if mustBeInPeriod && !status {
		t.Logf("status=%v : month=%s, day=%s, time=%s\n", status, statusMonth.String(), statusDay.String(), statusTime.String())
		return fmt.Errorf("%s should be in period %+v", ti, c.Periods)
	}
	if !mustBeInPeriod && status {
		t.Logf("status=%v : month=%s, day=%s, time=%s\n", status, statusMonth.String(), statusDay.String(), statusTime.String())
		return fmt.Errorf("%s should NOT be in period %+v", ti, c.Periods)
	}
	return nil
}

func TestCalendarSimpleInclude(t *testing.T) {
	// include only dayOfMonth == 1
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{{Included: true, DaysOfMonth: &dayInterval{From: 1, To: 1}}}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 2, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarSimpleExclude(t *testing.T) {
	// exclude only dayOfMonth == 1 (calendar is empty by default, no date should be valid)
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{{Included: false, DaysOfMonth: &dayInterval{From: 1, To: 1}}}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 2, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarEmpty(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 2, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarMultipleInclude(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: true, DaysOfMonth: &dayInterval{From: 1, To: 1}},
		{Included: true, DaysOfMonth: &dayInterval{From: 3, To: 3}},
		{Included: true, DaysOfMonth: &dayInterval{From: 5, To: 5}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 2, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 3, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 4, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 5, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 6, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarIncludeCombo(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: true, MonthsOfYear: &monthInterval{From: 1, To: 1}, DaysOfMonth: &dayInterval{From: 1, To: 1}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 2, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 2, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarIncludeCombo2(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: true, MonthsOfYear: &monthInterval{From: 1, To: 1}, DaysOfMonth: &dayInterval{From: 1, To: 1}, HoursOfDay: &hoursInterval{FromHour: 12, ToHour: 13}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 2, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 2, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarIncludeExclude(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: true, DaysOfMonth: &dayInterval{From: 1, To: 3}},
		{Included: false, DaysOfMonth: &dayInterval{From: 2, To: 4}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 2, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 3, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 4, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarIncludeExclude2(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: false, DaysOfMonth: &dayInterval{From: 2, To: 4}},
		{Included: true, DaysOfMonth: &dayInterval{From: 1, To: 3}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 2, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 3, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 4, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarIncludeMonthsExcludeDays(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: true, MonthsOfYear: &monthInterval{From: 1, To: 12}},
		{Included: false, DaysOfMonth: &dayInterval{From: 1, To: 31}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarExcludeDaysIncludeMonths(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: true, MonthsOfYear: &monthInterval{From: 1, To: 12}},
		{Included: false, DaysOfMonth: &dayInterval{From: 1, To: 31}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarIncludeDaysExcludeMonths(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: true, DaysOfMonth: &dayInterval{From: 1, To: 31}},
		{Included: false, MonthsOfYear: &monthInterval{From: 1, To: 12}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarIncludeDaysExcludeMonths2(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: true, DaysOfMonth: &dayInterval{From: 1, To: 31}},
		{Included: false, MonthsOfYear: &monthInterval{From: 1, To: 6}},
		{Included: true, MonthsOfYear: &monthInterval{From: 1, To: 1}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 3, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 7, 1, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
}

func TestCalendarExcludeAllIncludeSpecific1(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: false, MonthsOfYear: &monthInterval{From: 1, To: 12}, DaysOfMonth: &dayInterval{From: 1, To: 31}},
		{Included: false, HoursOfDay: &hoursInterval{FromHour: 0, ToHour: 24}},
		{Included: true, MonthsOfYear: &monthInterval{From: 1, To: 1}, DaysOfMonth: &dayInterval{From: 1, To: 1}, HoursOfDay: &hoursInterval{FromHour: 12, ToHour: 13}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
}

func TestCalendarExcludeAllIncludeSpecific2(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: false, MonthsOfYear: &monthInterval{From: 1, To: 12}, DaysOfMonth: &dayInterval{From: 1, To: 31}},
		{Included: false, HoursOfDay: &hoursInterval{FromHour: 0, ToHour: 24}},
		{Included: true, MonthsOfYear: &monthInterval{From: 1, To: 1}, DaysOfMonth: &dayInterval{From: 1, To: 1}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarExcludeAllIncludeSpecific3(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "UTC", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: false, MonthsOfYear: &monthInterval{From: 1, To: 12}, DaysOfMonth: &dayInterval{From: 1, To: 31}},
		{Included: false, HoursOfDay: &hoursInterval{FromHour: 0, ToHour: 24}},
		{Included: true, MonthsOfYear: &monthInterval{From: 1, To: 1}},
	}}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}

func TestCalendarWithTimezone(t *testing.T) {
	cal := Calendar{ID: 1, Name: "cal", Timezone: "Europe/Paris", UnionCalendarIDs: []int64{}, Periods: []Period{
		{Included: true, DaysOfMonth: &dayInterval{From: 2, To: 3}},
	}}

	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 23, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 1, 21, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}

	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 3, 21, 30, 0, 0, time.UTC), true); err != nil {
		t.Error(err)
	}
	if err := checkCalendarPeriod(t, cal, time.Date(2020, 1, 3, 23, 30, 0, 0, time.UTC), false); err != nil {
		t.Error(err)
	}
}
