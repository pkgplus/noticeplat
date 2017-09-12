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

	cur := time.Unix(1504850401, 0)
	next := s.NextRunTime(cur).Unix()
	if next != 1504854000 {
		t.Fatalf("expect 1504850760 but get %d", next)
	}

	s, err = newTestSetting_workDay()
	next = s.NextRunTime(cur).Unix()
	if next != 1505109600 {
		t.Fatalf("expect 1505109600 but get %d", next)
	}
}

func newTestSetting() (*Setting, error) {
	return NewSetting([]byte(`
{
"first":1504850400,
"Intervals": ["@every 1h"]
}
`))
}

func newTestSetting_workDay() (*Setting, error) {
	return NewSetting([]byte(`
{
"first":1504850400,
"Intervals": ["@every workday"]
}
`))
}
