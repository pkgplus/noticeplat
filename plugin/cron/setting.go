package cron

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
)

const every = "@every "

type Setting struct {
	// Clock     string // HH:mm
	First     int64
	Intervals []string // @every 1h | 100d

	firsttime time.Time
}

func NewSetting(setting_bytes []byte) (s *Setting, err error) {
	s = new(Setting)
	err = json.Unmarshal(setting_bytes, s)
	if err != nil {
		return nil, err
	}

	if len(s.Intervals) == 0 {
		return nil, errors.New("interval is required")
	}

	if s.First == 0 {
		return nil, errors.New("first time required")
		// clock, err := time.Parse("15:04", s.Clock)
		// if err != nil {
		// 	return nil, err
		// }
		// minduraion, err := s.getMinInterval()
		// if err != nil {
		// 	return nil, err

		// }

		// // if time.Now().After(clock) {
		// s.firsttime = clock.Add(minduraion)
		// s.First = s.firsttime.Unix()
		// // }
	} else {
		s.firsttime = time.Unix(s.First, 0)
	}

	return
}

func (s *Setting) getMinInterval() (minDuration time.Duration, err error) {
	for i, interval := range s.Intervals {
		if strings.HasPrefix(interval, every) {
			duration_desc := interval[len(every):]

			var duration time.Duration
			if duration_desc == "workday" {
				duration, err = time.ParseDuration("24h")
			} else {
				duration, err = time.ParseDuration(duration_desc)
			}
			if err != nil {
				return
			}

			if i == 0 || duration.Seconds() < minDuration.Seconds() {
				minDuration = duration
			}
		}
	}

	return
}

func (s *Setting) NextRunTime(curtime time.Time) (next_run time.Time) {
	if s.firsttime.IsZero() {
		s.firsttime = time.Unix(s.First, 0)
	}

	for i, interval := range s.Intervals {
		if strings.HasPrefix(interval, every) {
			var duration time.Duration
			var err error
			duration_desc := interval[len(every):]
			if duration_desc == "workday" {
				duration, err = time.ParseDuration("24h")
			} else {
				duration, err = time.ParseDuration(duration_desc)
			}
			if err != nil {
				log.Printf("parse interval setting %s failed:%v", interval, err)
				return
			}

			var next_run_this time.Time
			del := curtime.Sub(s.firsttime)
			loopedtimes := del / duration
			// if del.Seconds() > 0 && loopedtimes == 0 {
			next_run_this = s.firsttime.Add((loopedtimes + 1) * duration)
			// } else {
			// }

			// workday
			if duration_desc == "workday" {
				weekday := int(next_run_this.Weekday())
				if weekday == 6 {
					next_run_this = next_run_this.Add(48 * time.Hour)
				} else if weekday == 0 {
					next_run_this = next_run_this.Add(24 * time.Hour)
				}
			}
			if i == 0 || next_run_this.Sub(next_run).Nanoseconds() < 0 {
				next_run = next_run_this
			}
		}
	}

	return
}

func (s *Setting) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
