package cron

import (
	"testing"
	"time"
)

func TestNextRunTime(t *testing.T) {
	s, err := newTestSetting()
	if err != nil {
		t.Fatal(err)
	}

	// cur := time.Now().Truncate(time.Minute)
	cur := time.Unix(1505552769, 0).Truncate(time.Minute)
	next := s.NextRunTime(cur).Unix()
	if next != cur.Add(time.Minute).Unix() {
		t.Fatalf("expect %d but get %d", cur.Add(time.Minute).Unix(), next)
	}

	s2, err := newTestSetting_workDay()
	if err != nil {
		t.Fatal(err)
	}
	next = s2.NextRunTime(cur).Unix()
	if next != 1505692800 {
		t.Fatalf("expect 1505692800 but get %d", next)
	}
}

func newTestSetting() (*CronSetting, error) {
	return NewSetting([]byte(`
{
"firstTimeStr":"20170916103000",
"interval": "1m"
}
`))
}

func newTestSetting_workDay() (*CronSetting, error) {
	return NewSetting([]byte(`
{
"interval": "1m",
"weekLimit":"weekday",
"clockLimitStart":"08:00",
"clockLimitEnd":"12:00"
}
`))
}
