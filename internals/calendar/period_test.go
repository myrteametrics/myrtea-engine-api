package calendar

// func TestMarshalUnmarshal(t *testing.T) {

// 	// Thursday  2019-08-01
// 	date, _ := time.Parse("2006-01-02", "2019-08-01")
// 	p := Period{
// 		DateTimeIntervals: []dateTimeInterval{
// 			{
// 				From: date.AddDate(0, 0, -1),
// 				To:   date.AddDate(0, 0, 1),
// 			},
// 		},
// 		MonthsOfYear: []monthInterval{
// 			{
// 				From: time.January,
// 				To:   time.January,
// 			},
// 			{
// 				From: time.August,
// 				To:   time.August,
// 			},
// 		},
// 		DaysOfMonth: []dayInterval{
// 			{
// 				From: date.Day(),
// 				To:   date.Day() + 1,
// 			},
// 		},
// 		DaysOfWeek: []dayWeekInterval{
// 			{
// 				From: time.Monday,
// 				To:   time.Monday,
// 			},
// 			{
// 				From: time.Thursday,
// 				To:   time.Thursday,
// 			},
// 		},
// 		HoursOfDay: []hoursInterval{
// 			{
// 				FromHour:   10,
// 				FromMinute: 30,
// 				ToHour:     12,
// 				ToMinute:   30,
// 			},
// 			{
// 				FromHour:   14,
// 				FromMinute: 00,
// 				ToHour:     18,
// 				ToMinute:   00,
// 			},
// 		},
// 	}
// 	pData, _ := json.Marshal(p)

// 	var _p Period
// 	json.Unmarshal(pData, &_p)

// 	if fmt.Sprint(p) != fmt.Sprint(_p) {
// 		t.Error("error in Marshal -> Unmarshal")
// 	}
// }

// func TestPeriod(t *testing.T) {

// 	// Thursday  2019-08-01
// 	date, _ := time.Parse("2006-01-02", "2019-08-01")

// 	// period zero value
// 	p := Period{}
// 	if !p.inPeriod(date) {
// 		t.Error("inPeriod should return true")
// 	}

// 	// add a dates interval in period
// 	p.DateTimeIntervals = []dateTimeInterval{
// 		{
// 			From: date.AddDate(0, 0, -1),
// 			To:   date.AddDate(0, 0, 1),
// 		},
// 	}
// 	if !p.inPeriod(date) {
// 		t.Error("inPeriod should return true")
// 	}

// 	// add month of year in period
// 	p.MonthsOfYear = []monthInterval{
// 		{
// 			From: time.January,
// 			To:   time.January,
// 		},
// 	}
// 	if p.inPeriod(date) {
// 		t.Error("inPeriod should return false")
// 	}

// 	p.MonthsOfYear = append(p.MonthsOfYear,
// 		monthInterval{
// 			From: time.August,
// 			To:   time.August,
// 		},
// 	)
// 	if !p.inPeriod(date) {
// 		t.Error("inPeriod should return true")
// 	}

// 	// add days of month in period
// 	p.DaysOfMonth = []dayInterval{
// 		{
// 			From: date.Day() + 1,
// 			To:   date.Day() + 1,
// 		},
// 	}
// 	if p.inPeriod(date) {
// 		t.Error("inPeriod should return false")
// 	}

// 	p.DaysOfMonth = []dayInterval{
// 		{
// 			From: date.Day(),
// 			To:   date.Day() + 1,
// 		},
// 	}
// 	if !p.inPeriod(date) {
// 		t.Error("inPeriod should return true")
// 	}

// 	// add days of week in period
// 	p.DaysOfWeek = []dayWeekInterval{
// 		{
// 			From: time.Monday,
// 			To:   time.Monday,
// 		},
// 	}
// 	if p.inPeriod(date) {
// 		t.Error("inPeriod should return false")
// 	}
// 	p.DaysOfWeek = append(p.DaysOfWeek,
// 		dayWeekInterval{
// 			From: time.Thursday,
// 			To:   time.Thursday,
// 		},
// 	)
// 	if !p.inPeriod(date) {
// 		t.Error("inPeriod should return true")
// 	}

// 	// add hours of day in period
// 	p.HoursOfDay = []hoursInterval{
// 		{
// 			FromHour:   10,
// 			FromMinute: 30,
// 			ToHour:     12,
// 			ToMinute:   30,
// 		},
// 	}
// 	date = date.Add(time.Hour*14 + time.Minute*30)
// 	if p.inPeriod(date) {
// 		t.Error("inPeriod should return false")
// 	}
// 	p.HoursOfDay = []hoursInterval{
// 		{
// 			FromHour:   10,
// 			FromMinute: 30,
// 			ToHour:     12,
// 			ToMinute:   30,
// 		},
// 		{
// 			FromHour:   14,
// 			FromMinute: 00,
// 			ToHour:     18,
// 			ToMinute:   00,
// 		},
// 	}
// 	if !p.inPeriod(date) {
// 		t.Error("inPeriod should return true")
// 	}
// }
