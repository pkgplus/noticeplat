package cron

import (
// "github.com/robfig/cron"
)

var (
// crontab *cron.Cron = cron.New()
)

type Cron struct {
	Setting *Setting
	Handler func()
}

func (c *Cron) GetNextRunTime() int64 {
	return 0
}

func GetNextRunTime(setting string) int64 {
	return 0
}
